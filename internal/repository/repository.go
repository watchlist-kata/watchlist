package repository

import (
	"errors"
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
	AddToWatchlist(watchlist *GormWatchlist) error
	RemoveFromWatchlist(mediaID uint, userID uint) error
	GetWatchlist(userID uint) ([]GormWatchlist, error)
	CheckInWatchlist(mediaID uint, userID uint) (bool, error)
}

// PostgresRepository реализует WatchlistRepository для PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// AddToWatchlist добавляет медиа в список просмотра пользователя
func (r *PostgresRepository) AddToWatchlist(watchlist *GormWatchlist) error {
	// Сначала проверяем, существует ли уже такая запись
	exists, err := r.CheckInWatchlist(watchlist.MediaID, watchlist.UserID)
	if err != nil {
		return err
	}

	// Если запись уже существует, возвращаем ошибку
	if exists {
		return ErrDuplicateEntry
	}

	// Создаем новую запись
	return r.db.Create(watchlist).Error
}

// RemoveFromWatchlist удаляет медиа из списка просмотра пользователя
func (r *PostgresRepository) RemoveFromWatchlist(mediaID uint, userID uint) error {
	// Сначала проверяем, существует ли запись
	exists, err := r.CheckInWatchlist(mediaID, userID)
	if err != nil {
		return err
	}

	// Если записи не существует, возвращаем ошибку
	if !exists {
		return ErrRecordNotFound
	}

	// Удаляем запись
	return r.db.Where("media_id = ? AND user_id = ?", mediaID, userID).Delete(&GormWatchlist{}).Error
}

// GetWatchlist получает список просмотра пользователя
func (r *PostgresRepository) GetWatchlist(userID uint) ([]GormWatchlist, error) {
	var watchlists []GormWatchlist
	if err := r.db.Where("user_id = ?", userID).Find(&watchlists).Error; err != nil {
		return nil, err
	}
	return watchlists, nil
}

// CheckInWatchlist проверяет, находится ли медиа в списке просмотра пользователя
func (r *PostgresRepository) CheckInWatchlist(mediaID uint, userID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&GormWatchlist{}).Where("media_id = ? AND user_id = ?", mediaID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
