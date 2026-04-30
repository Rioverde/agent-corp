package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	authpb "github.com/Rioverde/agent-corp/api/proto/auth"
)

// Server implements auth.AuthServiceServer.
type Server struct {
	authpb.UnimplementedAuthServiceServer

	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewServer constructs an auth gRPC server.
func NewServer(pool *pgxpool.Pool, logger *zap.Logger) *Server {
	return &Server{pool: pool, logger: logger}
}

// Register is a stub returning a placeholder user_id. Replace with real logic.
func (s *Server) Register(_ context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	s.logger.Info("Register received", zap.String("email", req.GetEmail()))
	return &authpb.RegisterResponse{UserId: "stub-user-id"}, nil
}

// Login is a stub returning placeholder tokens. Replace with real logic.
func (s *Server) Login(_ context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	s.logger.Info("Login received", zap.String("email", req.GetEmail()))
	return &authpb.LoginResponse{
		AccessToken:  "stub-access-token",
		RefreshToken: "stub-refresh-token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}, nil
}
