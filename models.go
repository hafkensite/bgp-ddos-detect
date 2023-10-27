package main

import (
	"time"

	"gorm.io/gorm"
)

type TransitRoute struct {
	gorm.Model
	Timestamp     time.Time
	ISPAS         int
	TransitAS     int
	DestinationAS int
	Json          string
	Prefixes      string
}
