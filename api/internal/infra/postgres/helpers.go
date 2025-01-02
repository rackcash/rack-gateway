package postgres

import (
	"errors"

	"gorm.io/gorm"
)

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, gorm.ErrRecordNotFound)
}
