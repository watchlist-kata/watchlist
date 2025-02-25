package main

import (
	"fmt"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"

	"github.com/watchlist-kata/protos/watchlist"
	"github.com/watchlist-kata/watchlist/internal/config"
	"github.com/watchlist-kata/watchlist/internal/repository"
	"github.com/watchlist-kata/watchlist/internal/service"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация подключения к базе данных
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Автомиграция таблицы watchlist
	err = db.AutoMigrate(&repository.GormWatchlist{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Создание репозитория
	repo := repository.NewPostgresRepository(db)

	// Создание сервиса
	svc := service.NewWatchlistService(repo)

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	watchlist.RegisterWatchlistServiceServer(s, svc)

	log.Printf("Starting gRPC server on %s\n", cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
