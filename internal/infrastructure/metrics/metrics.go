package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	EmailSentTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "notification_server_email_sent_total",
			Help: "Total number of email sent",
		},
	)

	KafkaMessageProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "notification_service_kafka_message_processed_total",
			Help: "Total number of Kafka messages processed",
		},
	)
	OTPSentTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "notification_service_otp_send_total",
			Help: "Total number of OTP send ",
		},
	)
	EmailSendErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "notification_service_email_send_errors_total",
			Help: "Total number of email send errors",
		},
	)
)

func InitMetrics() {
	prometheus.MustRegister(EmailSentTotal)
	prometheus.MustRegister(EmailSendErrors)
	prometheus.MustRegister(KafkaMessageProcessed)
	prometheus.MustRegister(OTPSentTotal)
}

func StartMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":9090", nil)
	}()
}
