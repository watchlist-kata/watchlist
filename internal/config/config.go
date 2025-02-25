package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config содержит параметры конфигурации приложения
type Config struct {
	DBHost        string   // Хост базы данных
	DBPort        string   // Порт базы данных
	DBUser        string   // Пользователь базы данных
	DBPassword    string   // Пароль базы данных
	DBName        string   // Имя базы данных
	DBSSLMode     string   // Режим SSL для базы данных
	KafkaBrokers  []string // Список брокеров Kafka
	KafkaTopic    string   // Тема Kafka
	GRPCPort      string   // Порт для gRPC сервиса
	ServiceName   string   // Имя сервиса
	LogBufferSize int      // Размер буфера для логов
}

// LoadConfig загружает конфигурацию из .env файла
func LoadConfig() (*Config, error) {
	// Загружаем переменные окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	// Проверяем обязательные переменные окружения
	requiredEnvVars := []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD",
		"DB_NAME", "DB_SSLMODE", "KAFKA_BROKERS", "KAFKA_TOPIC",
		"GRPC_PORT", "SERVICE_NAME", "LOG_BUFFER_SIZE",
	}

	for _, envVar := range requiredEnvVars {
		if value := os.Getenv(envVar); value == "" {
			return nil, fmt.Errorf("missing required environment variable: %s", envVar)
		}
	}

	// Преобразуем KAFKA_BROKERS в []string
	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(kafkaBrokers) == 0 || (len(kafkaBrokers) == 1 && kafkaBrokers[0] == "") {
		return nil, fmt.Errorf("invalid KAFKA_BROKERS value")
	}

	// Преобразуем LOG_BUFFER_SIZE в int с дефолтным значением 100, если не задано корректно
	logBufferSize, err := strconv.Atoi(os.Getenv("LOG_BUFFER_SIZE"))
	if err != nil || logBufferSize <= 0 {
		logBufferSize = 100 // Значение по умолчанию
	}

	// Возвращаем конфигурацию
	return &Config{
		DBHost:        os.Getenv("DB_HOST"),
		DBPort:        os.Getenv("DB_PORT"),
		DBUser:        os.Getenv("DB_USER"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_NAME"),
		DBSSLMode:     os.Getenv("DB_SSLMODE"),
		KafkaBrokers:  kafkaBrokers,
		KafkaTopic:    os.Getenv("KAFKA_TOPIC"),
		GRPCPort:      os.Getenv("GRPC_PORT"),
		ServiceName:   os.Getenv("SERVICE_NAME"),
		LogBufferSize: logBufferSize,
	}, nil
}
