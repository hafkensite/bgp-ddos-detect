package main

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	err = db.AutoMigrate(&TransitRoute{})
	if err != nil {
		panic("failed to migrade TransitRoute database")
	}

	var subsciptions map[chan *TransitRoute]bool = make(map[chan *TransitRoute]bool)

	go func() {
		for {
			err := ingestPrefixAnnouncements(db, subsciptions)
			if err != nil {
				log.Print(err)
			}
		}
	}()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/config.js", configAPI)
	http.Handle("/api/ws", websocket.Handler(getBGPServer(db, subsciptions)))
	log.Fatal(http.ListenAndServe(":1337", nil))
}
