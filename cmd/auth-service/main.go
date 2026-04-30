package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	authpb "github.com/Rioverde/agent-corp/api/proto/auth"
	"github.com/Rioverde/agent-corp/internal/auth"
	"github.com/Rioverde/agent-corp/internal/config"
	"github.com/Rioverde/agent-corp/internal/db"
)

const configLoadTimeout = 20 * time.Second

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	ctx, cancel := context.WithTimeout(context.Background(), configLoadTimeout)
	defer cancel()

	cfg, err := config.NewConfig(ctx, logger)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	if err := db.RunMigrations(ctx, cfg.Database); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	pool, err := db.NewPool(context.Background(), cfg.Database)
	if err != nil {
		logger.Fatal("Failed to create database pool", zap.Error(err))
	}
	defer pool.Close()

	lis, err := net.Listen("tcp", cfg.GRPCServer.Address)
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	grpcSrv := grpc.NewServer()
	authpb.RegisterAuthServiceServer(grpcSrv, auth.NewServer(pool, logger))
	reflection.Register(grpcSrv)

	go func() {
		logger.Info("gRPC server listening", zap.String("address", cfg.GRPCServer.Address))
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Fatal("Serve failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down")
	grpcSrv.GracefulStop()
	logger.Info("Stopped")
}
