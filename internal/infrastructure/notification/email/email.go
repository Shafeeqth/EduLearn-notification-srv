package email

import (
	"context"
	"crypto/tls"
	"net/smtp"
	"sync"

	"github.com/Shafeeqth/notification-service/internal/domain"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/ratelimit"
	"github.com/jordan-wright/email"
	"go.uber.org/zap"
)

type EmailSender struct {
	smtpHost    string
	smtpPort    string
	username    string
	password    string
	logger      *zap.Logger
	pool        *smtpPool
	ratelimiter *ratelimit.RateLimiter
	emailPool   sync.Pool
}

type smtpPool struct {
	conns chan *smtp.Client
	mutex sync.Mutex
	host  string
	auth  smtp.Auth
}

func newSMTPPool(host, port, username, password string, size int) (*smtpPool, error) {
	pool := &smtpPool{
		conns: make(chan *smtp.Client, size),
		host:  host,
		auth:  smtp.PlainAuth("", username, password, host),
	}
	addr := host + ":" + port
	for i := 0; i < size; i++ {
		client, err := smtp.Dial(addr)
		if err != nil {
			return nil, err
		}

		// Upgrade the connection to use STARTTLS
		if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			client.Close()
			return nil, err
		}

		if err := client.Auth(pool.auth); err != nil {
			client.Close()
			return nil, err
		}
		pool.conns <- client
	}
	return pool, nil
}

func (p *smtpPool) Get() (*smtp.Client, error) {
	select {
	case client := <-p.conns:
		return client, nil
	default:
		client, err := smtp.Dial(p.host + ":587")
		if err != nil {
			return nil, err
		}

		// Upgrade the connection to use STARTTLS
		if err := client.StartTLS(&tls.Config{ServerName: p.host}); err != nil {
			client.Close()
			return nil, err
		}

		if err := client.Auth(p.auth); err != nil {
			client.Close()
			return nil, err
		}
		return client, nil
	}
}

func (p *smtpPool) Put(client *smtp.Client) {
	select {
	case p.conns <- client:
	default:
		client.Close()
	}
}

func NewEmailSender(smtpHost, smtpPort, username, password string, logger *zap.Logger) (*EmailSender, error) {

	pool, err := newSMTPPool(smtpHost, smtpPort, username, password, 5) // pool size of 5
	if err != nil {
		logger.Error("Failed to create SMTP pool", zap.Error(err))
		return nil, err
	}
	logger.Info("Connected to SMTP pool", zap.String("smtphost", smtpHost))

	return &EmailSender{
		smtpHost:    smtpHost,
		smtpPort:    smtpPort,
		username:    username,
		password:    password,
		logger:      logger,
		pool:        pool,
		ratelimiter: ratelimit.NewRateLimiter(10, 20), // Helpful to not tag emails as spam (10 emails/sec, burst of 20)
		emailPool: sync.Pool{
			New: func() interface{} {
				return email.NewEmail()
			},
		},
	}, nil
}

func (e *EmailSender) Send(ctx context.Context, notification domain.Notification) error {
	// Check rate limit
	if err := e.ratelimiter.Allow(ctx, notification.Recipient); err != nil {
		e.logger.Warn("Rate limit exceeded for email sending",
			zap.String("recipient", notification.Recipient))
		return err
	}
	e.logger.Info("Inside Send method (:)")

	// Get email message from pool
	msg := e.emailPool.Get().(*email.Email)
	defer func() {
		msg.From = ""
		msg.To = nil
		msg.Subject = ""
		msg.Text = nil
		msg.HTML = nil
		e.emailPool.Put(msg)
	}()
	msg.From = e.username
	msg.To = []string{notification.Recipient}
	msg.Subject = notification.Subject
	msg.HTML = []byte(notification.Body) // Use HTML field for HTML content

	client, err := e.pool.Get()
	if err != nil {
		e.logger.Error("Failed to get SMTP client from pool", zap.Error(err))
		return err
	}
	/*
		defer e.pool.Put(client)

		// msg := email.NewEmail()
		// msg.From = e.username
		// msg.To = []string{recipient}
		// msg.Subject = subject
		// msg.Text = []byte(body)
	*/
	defer func() {
		if err := client.Quit(); err != nil {
			e.logger.Warn("Failed to quit SMTP client", zap.Error(err))
		}
		e.pool.Put(client)
	}()

	// Start the SMTP mail transaction
	if err := client.Mail(e.username); err != nil {
		e.logger.Error("Failed to start mail transaction", zap.Error(err))
		return err
	}
	for _, addr := range msg.To {
		if err := client.Rcpt(addr); err != nil {
			e.logger.Error("Failed to add recipient", zap.String("recipient", addr), zap.Error(err))
			return err
		}
	}

	// Write the email data to the SMTP client
	writer, err := client.Data()
	if err != nil {
		e.logger.Error("Failed to get SMTP data writer", zap.Error(err))
		return err
	}
	data, err := msg.Bytes()
	if err != nil {
		e.logger.Error("Failed to generate email bytes", zap.Error(err))
		return err
	}
	if _, err := writer.Write(data); err != nil {
		e.logger.Error("Failed to write email data", zap.Error(err))
		return err
	}
	if err := writer.Close(); err != nil {
		e.logger.Error("Failed to close SMTP data writer", zap.Error(err))
		return err
	}

	e.logger.Info("Email sent successfully", zap.String("recipient", notification.Recipient))
	return nil
}
