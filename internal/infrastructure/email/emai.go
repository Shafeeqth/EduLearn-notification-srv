// package email

// import (
// 	"net/smtp"

// 	"github.com/jordan-wright/email"
// 	"go.uber.org/zap"
// )

// type EmailSender struct {
// 	smtpHost string
// 	smtpPort string
// 	username string
// 	password string
// 	logger   *zap.Logger
// }

// func NewEmailSender(smtpHost, smtpPort, username, password string, logger *zap.Logger) *EmailSender {

// 	return &EmailSender{
// 		smtpHost: smtpHost,
// 		smtpPort: smtpPort,
// 		username: username,
// 		password: password,
// 		logger:   logger,
// 	}
// }

// func (e *EmailSender) SendEmail(recipient, subject, body string) error {
// 	msg := email.NewEmail()
// 	msg.From = e.username
// 	msg.To = []string{recipient}
// 	msg.Subject = subject
// 	msg.Text = []byte(body)

// 	addr := e.smtpHost + ":" + e.smtpPort
// 	auth := smtp.PlainAuth("", e.username, e.password, e.smtpHost)

// 	if err := msg.Send(addr, auth); err != nil {
// 		e.logger.Error("Failed to send email",
// 			zap.String("recipient", recipient),
// 			zap.Error(err),
// 		)
// 		return err
// 	}
// 	e.logger.Info("Email send", zap.String("recipient", recipient))
// 	return nil
// }
