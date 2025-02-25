package repository

import (
	"time"
)

// GormWatchlist представляет модель списка просмотра в базе данных
type GormWatchlist struct {
	ID        uint `gorm:"primaryKey"`
	MediaID   uint
	UserID    uint
	CreatedAt time.Time
}

// TableName возвращает имя таблицы для модели GormWatchlist
func (GormWatchlist) TableName() string {
	return "watchlist"
}
