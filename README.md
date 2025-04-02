# Notification Service

The **Notification Service** is a microservice designed to handle notifications (email and in-app) and OTP (One-Time Password) generation and delivery. It is built using **Go** and leverages technologies like **Kafka**, **Redis**, **gRPC**, and **Prometheus** for scalability, performance, and monitoring.

---

## Project Structure

Follows a clean architecture pattern, separating concerns into distinct layers:

```
notification-service/
├── cmd/
│   └── server/
│       └── main.go          # Entry point
├── internal/
│   ├── domain/
│   │   ├── notification.go  # Domain entities and interfaces
│   │   ├── otp.go           # OTP-related domain logic
│   │   └── errors.go        # Custom error types
│   ├── application/
│   │   ├── service/
│   │   │   ├── notification_service.go # Business logic for notifications
│   │   │   └── otp_service.go          # Business logic for OTP
│   │   └── dto/
│   │       └── notification_dto.go     # Data transfer objects
│   ├── infrastructure/
│   │   ├── kafka/
│   │   │   ├── producer.go  # Kafka producer
│   │   │   └── consumer.go  # Kafka consumer with worker pool
│   │   ├── notification/
│   │   │   ├── email/
│   │   │   │   └── email.go     # Email sending
│   │   │   └── strategy.go      # Notification strategy interface
│   │   ├── database/
│   │   │   ├── gorm.go      # GORM database connection
│   │   │   └── repository.go # GORM repository
│   │   ├── redis/
│   │   │   └── redis.go     # Redis client for OTP storage
│   │   ├── config/
│   │   │   └── config.go    # Configuration management
│   │   ├── logging/
│   │   │   └── logger.go    # Logging setup
│   │   ├── metrics/
│   │   │   └── metrics.go   # Prometheus metrics
│   │   └── ratelimit/
│   │       └── ratelimit.go # Rate limiting
│   ├── presentation/
│   │   ├── grpc/
│   │   │   ├── server.go    # gRPC server
│   │   │   └── handler.go   # gRPC handlers
│   │   └── proto/
│   │       └── notification.proto # gRPC proto file
├── go.mod
└── go.sum
```

---

## Tech Stack

The Notification Service is built using the following technologies:

- **Programming Language**: Go (Golang)
- **Message Broker**: Kafka (for asynchronous processing)
- **Database**: PostgreSQL (for persistent storage)
- **Cache**: Redis (for OTP storage and caching)
- **API**: gRPC (for high-performance communication)
- **Metrics**: Prometheus (for monitoring and alerting)
- **Logging**: Zap (structured logging)
- **Dependency Management**: Go Modules
- **Containerization**: Docker (for running dependencies and the service)
- **Build Tools**: Protobuf Compiler (for gRPC code generation)

---

## Features

- **Email Notifications**: Send email notifications using SMTP.
- **In-App Notifications**: Handle in-app notifications (future implementation).
- **OTP Management**: Generate, store, and validate OTPs using Redis.
- **Kafka Integration**: Use Kafka for asynchronous notification processing.
- **Rate Limiting**: Prevent spamming by limiting email sending rates.
- **Prometheus Metrics**: Expose metrics for monitoring and alerting.
- **gRPC API**: Provide a gRPC interface for external communication.

---

## Prerequisites

- **Go**: Version 1.18 or higher.
- **Docker**: For running Kafka, Redis, and other dependencies.
- **Protobuf Compiler**: For generating gRPC code.

---

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/your-repo/notification-service.git
   cd notification-service
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Generate gRPC code:
   ```bash
   protoc --go_out=. --go-grpc_out=. internal/presentation/proto/notification.proto
   ```

4. Start dependencies using Docker Compose:
   ```bash
   docker-compose up -d
   ```

---

## Configuration

The service uses a configuration file (`config.yaml`) or environment variables. Key configurations include:

- **Database**:
  - `DATABASE_DSN`: PostgreSQL connection string.
- **Redis**:
  - `REDIS_ADDR`: Redis server address.
- **Kafka**:
  - `KAFKA_BROKERS`: Kafka broker addresses.
- **SMTP**:
  - `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`: SMTP server details.

---

## Running the Service

Start the service:
```bash
go run cmd/server/main.go
```

---

## API Endpoints

### gRPC Endpoints

1. **Send OTP**:
   - **Method**: `SendOTP`
   - **Request**:
     ```json
     {
       "userId": "123",
       "email": "user@example.com"
     }
     ```
   - **Response**:
     ```json
     {
       "success": true,
       "message": "OTP sent successfully"
     }
     ```

2. **Get All Notifications**:
   - **Method**: `GetAllNotifications`
   - **Request**:
     ```json
     {
       "userId": "123"
     }
     ```
   - **Response**:
     ```json
     {
       "notifications": [
         {
           "id": "1",
           "type": "email",
           "subject": "Welcome",
           "body": "Welcome to EduLearn!",
           "isRead": false,
           "createdAt": "2023-03-24T12:00:00Z"
         }
       ]
     }
     ```

---

## Monitoring

Prometheus metrics are exposed at `/metrics` (default port: `9090`).

---

## Testing

Run unit tests:
```bash
go test ./...
```

---

## Contributing

1. Fork the repository.
2. Create a feature branch.
3. Submit a pull request.

---

## License

This project is licensed under the MIT License.