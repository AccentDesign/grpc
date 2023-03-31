package db

import (
	"gorm.io/gorm"
)

func Connect(dialector gorm.Dialector, config gorm.Config) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, &config)
	if err != nil {
		return nil, err
	}

	return db, nil
}
