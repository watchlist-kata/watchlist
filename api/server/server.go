package server

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/watchlist-kata/protos/watchlist"
	"github.com/watchlist-kata/watchlist/internal/config"
	"github.com/watchlist-kata/watchlist/internal/repository"
	"github.com/watchlist-kata/watchlist/internal/service"
	"github.com/watchlist-kata/watchlist/pkg/utils"
	"google.golang.org/grpc"
)

// RunServer запускает gRPC сервер
func RunServer(cfg *config.Config, logger *slog.Logger) error {
	// Подключение к базе данных
	db, err := utils.ConnectToDatabase(cfg)
	if err != nil {
		logger.Error("failed to connect to database", slog.Any("error", err))
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Создание репозитория
	repo := repository.NewPostgresRepository(db, logger)

	// Создание сервиса
	svc := service.NewWatchlistService(repo, logger)

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen", slog.Any("error", err))
		return fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()
	watchlist.RegisterWatchlistServiceServer(s, svc)

	logger.Info("starting gRPC server", slog.String("port", cfg.GRPCPort))
	fmt.Printf("Starting gRPC server on %s\n", cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		logger.Error("failed to serve", slog.Any("error", err))
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}
