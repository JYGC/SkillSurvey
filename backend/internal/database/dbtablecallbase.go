package database

import (
	"gorm.io/gorm"
)

type DbTableCallBase struct {
	db *gorm.DB
}

func (d DbTableCallBase) MigrateTable(entityType interface{}) {
	if !d.db.Migrator().HasTable(entityType) {
		d.db.Migrator().CreateTable(entityType)
	}
	// Add ability to chnage columns
}
