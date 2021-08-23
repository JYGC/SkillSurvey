package database

import "github.com/HouzuoGuo/tiedot/db"

type CollectionBase struct {
	database       db.DB
	CollectionName string
}
