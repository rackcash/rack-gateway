package domain

import "time"

type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
}
