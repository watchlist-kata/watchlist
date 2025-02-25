package service

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/watchlist-kata/protos/watchlist"
	"github.com/watchlist-kata/watchlist/internal/repository"
)

// WatchlistService реализует интерфейс сервиса WatchlistService из proto-файла
type WatchlistService struct {
	watchlist.UnimplementedWatchlistServiceServer
	repo repository.WatchlistRepository
}

// NewWatchlistService создает новый экземпляр WatchlistService
func NewWatchlistService(repo repository.WatchlistRepository) *WatchlistService {
	return &WatchlistService{repo: repo}
}

// AddToWatchlist добавляет медиа в список просмотра пользователя
func (s *WatchlistService) AddToWatchlist(ctx context.Context, req *watchlist.AddToWatchlistRequest) (*watchlist.AddToWatchlistResponse, error) {
	// Проверка входных данных
	if req.MediaId <= 0 || req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "media_id и user_id должны быть положительными числами")
	}

	watchlistItem := &repository.GormWatchlist{
		MediaID:   uint(req.MediaId),
		UserID:    uint(req.UserId),
		CreatedAt: time.Now(),
	}

	err := s.repo.AddToWatchlist(watchlistItem)
	if err != nil {
		// Если элемент уже существует, возвращаем успех (идемпотентность операции)
		if errors.Is(err, repository.ErrDuplicateEntry) {
			return &watchlist.AddToWatchlistResponse{Success: true}, nil
		}
		return nil, status.Errorf(codes.Internal, "ошибка при добавлении в watchlist: %v", err)
	}

	return &watchlist.AddToWatchlistResponse{Success: true}, nil
}

// RemoveFromWatchlist удаляет медиа из списка просмотра пользователя
func (s *WatchlistService) RemoveFromWatchlist(ctx context.Context, req *watchlist.RemoveFromWatchlistRequest) (*watchlist.RemoveFromWatchlistResponse, error) {
	// Проверка входных данных
	if req.MediaId <= 0 || req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "media_id и user_id должны быть положительными числами")
	}

	err := s.repo.RemoveFromWatchlist(uint(req.MediaId), uint(req.UserId))
	if err != nil {
		// Если элемент не найден, возвращаем соответствующее сообщение
		if errors.Is(err, repository.ErrRecordNotFound) {
			return &watchlist.RemoveFromWatchlistResponse{Success: false}, nil
		}
		return nil, status.Errorf(codes.Internal, "ошибка при удалении из watchlist: %v", err)
	}

	return &watchlist.RemoveFromWatchlistResponse{Success: true}, nil
}

// GetWatchlist получает список просмотра пользователя
func (s *WatchlistService) GetWatchlist(ctx context.Context, req *watchlist.GetWatchlistRequest) (*watchlist.GetWatchlistResponse, error) {
	// Проверка входных данных
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id должен быть положительным числом")
	}

	gormWatchlists, err := s.repo.GetWatchlist(uint(req.UserId))
	if err != nil {
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

	return &watchlist.GetWatchlistResponse{Watchlists: watchlistItems}, nil
}

// CheckInWatchlist проверяет, находится ли медиа в списке просмотра пользователя
func (s *WatchlistService) CheckInWatchlist(ctx context.Context, req *watchlist.CheckInWatchlistRequest) (*watchlist.CheckInWatchlistResponse, error) {
	// Проверка входных данных
	if req.MediaId <= 0 || req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "media_id и user_id должны быть положительными числами")
	}

	inWatchlist, err := s.repo.CheckInWatchlist(uint(req.MediaId), uint(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка при проверке наличия в watchlist: %v", err)
	}

	return &watchlist.CheckInWatchlistResponse{InWatchlist: inWatchlist}, nil
}
