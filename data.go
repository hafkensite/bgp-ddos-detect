package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"
	"gorm.io/gorm"
)

type AS struct {
	Num  int
	Name string
}

type DataRow struct {
	Timestamp     time.Time
	ISPAS         AS
	TransitAS     AS
	DestinationAS AS
	Prefix        string
	PrefixWhois   string
}

func getBGPServer(db *gorm.DB, subsciptions map[chan *TransitRoute]bool) func(ws *websocket.Conn) {

	return func(ws *websocket.Conn) {
		q := ws.Request().URL.Query().Get("asnr")
		messages := []TransitRoute{}
		asnr, _ := strconv.Atoi(q)

		oldest := time.Now().Add(-24 * time.Hour)

		liveFeed := make(chan *TransitRoute)
		subsciptions[liveFeed] = true
		defer delete(subsciptions, liveFeed)

		db.Where(fmt.Sprintf("timestamp >= datetime(%d, 'unixepoch')", oldest.Unix())).Find(&messages, &TransitRoute{TransitAS: asnr})

		for _, msg := range messages {
			msgToWS(&msg, ws)
		}

		for msg := range liveFeed {
			if msg.TransitAS == asnr {
				msgToWS(msg, ws)
			}
		}
	}
}

func msgToWS(msg *TransitRoute, ws *websocket.Conn) {
	ispName, _ := GetNameForAs(msg.ISPAS)
	transitName, _ := GetNameForAs(msg.TransitAS)
	destName, _ := GetNameForAs(msg.DestinationAS)

	for _, prefix := range strings.Split(msg.Prefixes, ",") {
		whois, err := GetNameForNet(prefix)
		if err != nil {
			panic(err)
		}
		row := DataRow{
			Timestamp:     msg.Timestamp,
			ISPAS:         AS{Num: msg.ISPAS, Name: ispName},
			TransitAS:     AS{Num: msg.TransitAS, Name: transitName},
			DestinationAS: AS{Num: msg.DestinationAS, Name: destName},
			Prefix:        prefix,
			PrefixWhois:   whois,
		}

		bytes, err := json.Marshal(row)
		if err != nil {
			panic(err)
		}
		ws.Write(bytes)
	}
}
