package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/watchlist-kata/protos/watchlist"
	"github.com/watchlist-kata/watchlist/internal/repository"
)

// WatchlistService реализует интерфейс сервиса WatchlistService из proto-файла
type WatchlistService struct {
	watchlist.UnimplementedWatchlistServiceServer
	repo   repository.WatchlistRepository
	logger *slog.Logger
}

// NewWatchlistService создает новый экземпляр WatchlistService
func NewWatchlistService(repo repository.WatchlistRepository, logger *slog.Logger) *WatchlistService {
	return &WatchlistService{repo: repo, logger: logger}
}

// checkContextCancelled проверяет отмену контекста и логирует ошибку
func (s *WatchlistService) checkContextCancelled(ctx context.Context, method string) error {
	select {
	case <-ctx.Done():
		s.logger.ErrorContext(ctx, fmt.Sprintf("%s operation canceled", method), slog.Any("error", ctx.Err()))
		return ctx.Err()
	default:
		return nil
	}
}

// AddToWatchlist добавляет медиа в список просмотра пользователя
func (s *WatchlistService) AddToWatchlist(ctx context.Context, req *watchlist.AddToWatchlistRequest) (*watchlist.AddToWatchlistResponse, error) {
	if err := s.checkContextCancelled(ctx, "AddToWatchlist"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	// Проверка входных данных
	if req.MediaId <= 0 || req.UserId <= 0 {
		s.logger.WarnContext(ctx, "invalid media_id or user_id: must be positive integers")
		return nil, status.Error(codes.InvalidArgument, "media_id и user_id должны быть положительными числами")
	}

	watchlistItem := &repository.GormWatchlist{
		MediaID:   uint(req.MediaId),
		UserID:    uint(req.UserId),
		CreatedAt: time.Now(),
	}

	err := s.repo.AddToWatchlist(ctx, watchlistItem)
	if err != nil {
		// Если элемент уже существует, возвращаем успех (идемпотентность операции)
		if errors.Is(err, repository.ErrDuplicateEntry) {
			s.logger.InfoContext(ctx, fmt.Sprintf("media already in watchlist for media ID: %d and user ID: %d", req.MediaId, req.UserId))
			return &watchlist.AddToWatchlistResponse{Success: true}, nil
		}
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to add media to watchlist for media ID: %d and user ID: %d", req.MediaId, req.UserId), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "ошибка при добавлении в watchlist: %v", err)
	}

	s.logger.InfoContext(ctx, fmt.Sprintf("media added to watchlist successfully for media ID: %d and user ID: %d", req.MediaId, req.UserId))
	return &watchlist.AddToWatchlistResponse{Success: true}, nil
}

// RemoveFromWatchlist удаляет медиа из списка просмотра пользователя
func (s *WatchlistService) RemoveFromWatchlist(ctx context.Context, req *watchlist.RemoveFromWatchlistRequest) (*watchlist.RemoveFromWatchlistResponse, error) {
	if err := s.checkContextCancelled(ctx, "RemoveFromWatchlist"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	// Проверка входных данных
	if req.MediaId <= 0 || req.UserId <= 0 {
		s.logger.WarnContext(ctx, "invalid media_id or user_id: must be positive integers")
		return nil, status.Error(codes.InvalidArgument, "media_id и user_id должны быть положительными числами")
	}

	err := s.repo.RemoveFromWatchlist(ctx, uint(req.MediaId), uint(req.UserId))
	if err != nil {
		// Если элемент не найден, возвращаем соответствующее сообщение
		if errors.Is(err, repository.ErrRecordNotFound) {
			s.logger.WarnContext(ctx, fmt.Sprintf("media not found in watchlist for media ID: %d and user ID: %d", req.MediaId, req.UserId))
			return &watchlist.RemoveFromWatchlistResponse{Success: false}, nil
		}
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to remove media from watchlist for media ID: %d and user ID: %d", req.MediaId, req.UserId), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "ошибка при удалении из watchlist: %v", err)
	}

	s.logger.InfoContext(ctx, fmt.Sprintf("media removed from watchlist successfully for media ID: %d and user ID: %d", req.MediaId, req.UserId))
	return &watchlist.RemoveFromWatchlistResponse{Success: true}, nil
}

// GetWatchlist получает список просмотра пользователя
func (s *WatchlistService) GetWatchlist(ctx context.Context, req *watchlist.GetWatchlistRequest) (*watchlist.GetWatchlistResponse, error) {
	if err := s.checkContextCancelled(ctx, "GetWatchlist"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	// Проверка входных данных
	if req.UserId <= 0 {
		s.logger.WarnContext(ctx, "invalid user_id: must be a positive integer")
		return nil, status.Error(codes.InvalidArgument, "user_id должен быть положительным числом")
	}

	gormWatchlists, err := s.repo.GetWatchlist(ctx, uint(req.UserId))
	if err != nil {
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to get watchlist for user ID: %d", req.UserId), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "ошибка при получении watchlist: %v", err)
	}

	watchlistItems := make([]*watchlist.WatchlistItem, 0, len(gormWatchlists))
	for _, gw := range gormWatchlists {
		watchlistItems = append(watchlistItems, &watchlist.WatchlistItem{
			Id:        int64(gw.ID),
			MediaId:   int64(gw.MediaID),
			UserId:    int64(gw.UserID),
			CreatedAt: gw.CreatedAt.Format(time.RFC3339),
		})
	}

	s.logger.InfoContext(ctx, fmt.Sprintf("watchlist fetched successfully for user ID: %d", req.UserId))
	return &watchlist.GetWatchlistResponse{Watchlists: watchlistItems}, nil
}

// CheckInWatchlist проверяет, находится ли медиа в списке просмотра пользователя
func (s *WatchlistService) CheckInWatchlist(ctx context.Context, req *watchlist.CheckInWatchlistRequest) (*watchlist.CheckInWatchlistResponse, error) {
	if err := s.checkContextCancelled(ctx, "CheckInWatchlist"); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	// Проверка входных данных
	if req.MediaId <= 0 || req.UserId <= 0 {
		s.logger.WarnContext(ctx, "invalid media_id or user_id: must be positive integers")
		return nil, status.Error(codes.InvalidArgument, "media_id и user_id должны быть положительными числами")
	}

	inWatchlist, err := s.repo.CheckInWatchlist(ctx, uint(req.MediaId), uint(req.UserId))
	if err != nil {
		s.logger.ErrorContext(ctx, fmt.Sprintf("failed to check media in watchlist for media ID: %d and user ID: %d", req.MediaId, req.UserId), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "ошибка при проверке наличия в watchlist: %v", err)
	}

	s.logger.InfoContext(ctx, fmt.Sprintf("media checked in watchlist for media ID: %d and user ID: %d", req.MediaId, req.UserId))
	return &watchlist.CheckInWatchlistResponse{InWatchlist: inWatchlist}, nil
}
