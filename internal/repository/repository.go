package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

var (
	// ErrRecordNotFound возвращается, когда запись не найдена
	ErrRecordNotFound = errors.New("record not found")
	// ErrDuplicateEntry возвращается при попытке создать дублирующуюся запись
	ErrDuplicateEntry = errors.New("duplicate entry")
)

// WatchlistRepository представляет интерфейс репозитория для работы со списками просмотра
type WatchlistRepository interface {
	AddToWatchlist(ctx context.Context, watchlist *GormWatchlist) error
	RemoveFromWatchlist(ctx context.Context, mediaID uint, userID uint) error
	GetWatchlist(ctx context.Context, userID uint) ([]GormWatchlist, error)
	CheckInWatchlist(ctx context.Context, mediaID uint, userID uint) (bool, error)
}

// PostgresRepository реализует WatchlistRepository для PostgreSQL
type PostgresRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(db *gorm.DB, logger *slog.Logger) *PostgresRepository {
	return &PostgresRepository{db: db, logger: logger}
}

// AddToWatchlist добавляет медиа в список просмотра пользователя
func (r *PostgresRepository) AddToWatchlist(ctx context.Context, watchlist *GormWatchlist) error {
	// Проверка отмены контекста
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, fmt.Sprintf("AddToWatchlist operation canceled for media ID: %d and user ID: %d", watchlist.MediaID, watchlist.UserID), slog.Any("error", ctx.Err()))
		return ctx.Err()
	default:
	}

	// Сначала проверяем, существует ли уже такая запись
	exists, err := r.CheckInWatchlist(ctx, watchlist.MediaID, watchlist.UserID)
	if err != nil {
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to check watchlist existence for media ID: %d and user ID: %d", watchlist.MediaID, watchlist.UserID), slog.Any("error", err))
		return err
	}

	// Если запись уже существует, возвращаем ошибку
	if exists {
		r.logger.WarnContext(ctx, fmt.Sprintf("media already in watchlist for media ID: %d and user ID: %d", watchlist.MediaID, watchlist.UserID))
		return ErrDuplicateEntry
	}

	// Создаем новую запись
	if err := r.db.Create(watchlist).Error; err != nil {
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to add media to watchlist for media ID: %d and user ID: %d", watchlist.MediaID, watchlist.UserID), slog.Any("error", err))
		return err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("media added to watchlist successfully for media ID: %d and user ID: %d", watchlist.MediaID, watchlist.UserID))
	return nil
}

// RemoveFromWatchlist удаляет медиа из списка просмотра пользователя
func (r *PostgresRepository) RemoveFromWatchlist(ctx context.Context, mediaID uint, userID uint) error {
	// Проверка отмены контекста
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, fmt.Sprintf("RemoveFromWatchlist operation canceled for media ID: %d and user ID: %d", mediaID, userID), slog.Any("error", ctx.Err()))
		return ctx.Err()
	default:
	}

	// Сначала проверяем, существует ли запись
	exists, err := r.CheckInWatchlist(ctx, mediaID, userID)
	if err != nil {
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to check watchlist existence for media ID: %d and user ID: %d", mediaID, userID), slog.Any("error", err))
		return err
	}

	// Если записи не существует, возвращаем ошибку
	if !exists {
		r.logger.WarnContext(ctx, fmt.Sprintf("media not found in watchlist for media ID: %d and user ID: %d", mediaID, userID))
		return ErrRecordNotFound
	}

	// Удаляем запись
	if err := r.db.Where("media_id = ? AND user_id = ?", mediaID, userID).Delete(&GormWatchlist{}).Error; err != nil {
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to remove media from watchlist for media ID: %d and user ID: %d", mediaID, userID), slog.Any("error", err))
		return err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("media removed from watchlist successfully for media ID: %d and user ID: %d", mediaID, userID))
	return nil
}

// GetWatchlist получает список просмотра пользователя
func (r *PostgresRepository) GetWatchlist(ctx context.Context, userID uint) ([]GormWatchlist, error) {
	// Проверка отмены контекста
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, fmt.Sprintf("GetWatchlist operation canceled for user ID: %d", userID), slog.Any("error", ctx.Err()))
		return nil, ctx.Err()
	default:
	}

	var watchlists []GormWatchlist
	if err := r.db.Where("user_id = ?", userID).Find(&watchlists).Error; err != nil {
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to get watchlist for user ID: %d", userID), slog.Any("error", err))
		return nil, err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("watchlist fetched successfully for user ID: %d", userID))
	return watchlists, nil
}

// CheckInWatchlist проверяет, находится ли медиа в списке просмотра пользователя
func (r *PostgresRepository) CheckInWatchlist(ctx context.Context, mediaID uint, userID uint) (bool, error) {
	// Проверка отмены контекста
	select {
	case <-ctx.Done():
		r.logger.ErrorContext(ctx, fmt.Sprintf("CheckInWatchlist operation canceled for media ID: %d and user ID: %d", mediaID, userID), slog.Any("error", ctx.Err()))
		return false, ctx.Err()
	default:
	}

	var count int64
	if err := r.db.Model(&GormWatchlist{}).Where("media_id = ? AND user_id = ?", mediaID, userID).Count(&count).Error; err != nil {
		r.logger.ErrorContext(ctx, fmt.Sprintf("failed to check media in watchlist for media ID: %d and user ID: %d", mediaID, userID), slog.Any("error", err))
		return false, err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("media checked in watchlist for media ID: %d and user ID: %d", mediaID, userID))
	return count > 0, nil
}
