package grpc

import (
	"net"

	"github.com/Shafeeqth/notification-service/internal/application/service"
	"github.com/Shafeeqth/notification-service/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	logger     *zap.Logger
}

func NewServer(notificationService *service.NotificationService, otpService *service.OTPService, logger *zap.Logger) *Server {
	grpcServer := grpc.NewServer()
	handler := NewHandler(notificationService, otpService, logger)
	proto.RegisterNotificationServiceServer(grpcServer, handler)
	return &Server{
		grpcServer: grpcServer, logger: logger,
	}

}

func (s *Server) Start(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		s.logger.Error("Failed to listen", zap.String("address", address), zap.Error(err))
		return err
	}
	s.logger.Info("gRPC server started", zap.String("address", address))
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
	s.logger.Info("gRPC server stopped")
}
