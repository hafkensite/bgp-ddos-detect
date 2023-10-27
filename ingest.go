package main

import (
	"encoding/json"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	rislive "github.com/a16/go-rislive/pkg/message"
	"golang.org/x/net/websocket"
	"gorm.io/gorm"
)

func ingestPrefixAnnouncements(db *gorm.DB, subscriptions map[chan *TransitRoute]bool) error {
	ws, err := websocket.Dial("wss://ris-live.ripe.net/v1/ws/?client=js-example-1", "", "https://ris-live.ripe.net/")
	if err != nil {
		return err
	}

	for transitAS, transitName := range transitNetworks {
		log.Printf("Subscribing to transit network %s (AS%d)", transitName, transitAS)
		filter := rislive.NewFilter()
		filter.Prefix = "0.0.0.0/0"
		filter.Path = strconv.Itoa(transitAS)
		subscribe := rislive.NewRisSubscribe(filter)
		err := websocket.JSON.Send(ws, subscribe)
		if err != nil {
			return err
		}
	}

	for {
		var msg rislive.RisLiveMessage
		err := websocket.JSON.Receive(ws, &msg)
		if err != nil {
			return err
		}

		switch msg.Type {
		case "ris_error":
			risErr := msg.Data.(*rislive.RisError)
			log.Printf("ris_error: %v, %v", risErr.CommandType, risErr.Message)
		case "ris_message":
			switch msg.BgpMsgType {
			case "UPDATE":
				risMsgUpdate := msg.Data.(*rislive.RisMessageUpdate)

				if risMsgUpdate.Origin != "IGP" && risMsgUpdate.Origin != "EGP" {
					continue
				}

				if len(risMsgUpdate.Path) <= 2 {
					continue
				}

				transitAS := -1
				path := make([]int, len(risMsgUpdate.Path))
				for i := range risMsgUpdate.Path {
					path[i], err = strconv.Atoi(string(risMsgUpdate.Path[i]))
					if err != nil {
						return err
					}
					_, ok := transitNetworks[path[i]]
					if ok {
						transitAS = path[i]
					}
				}

				// We don't want the transitAS to be the destination
				lastPath := path[len(path)-1]
				_, ok := transitNetworks[lastPath]
				if ok {
					// fmt.Print(",")
					continue
				}

				// Some transits have a continuous flood of traffic for a certain AS
				_, ok = denylistNetworks[lastPath]
				if ok {
					continue
				}

				// We only care about announcements from tier1 ISPs
				isp, ok := tier1Networks[path[0]]
				if !ok {
					// fmt.Printf("non-transit: %d\n", path[0])
					continue
				}

				sec, dec := math.Modf(risMsgUpdate.Timestamp)
				x, err := json.Marshal(risMsgUpdate)
				if err != nil {
					return err
				}

				prefixes := []string{}
				for _, announcement := range risMsgUpdate.Announcements {
					for _, prefix := range announcement.Prefixes {
						log.Printf("prefix %s seen via %s -> %s -> %d", prefix, isp, transitNetworks[transitAS], lastPath)
						prefixes = append(prefixes, prefix)
					}

				}
				// log.Printf("ris_message(UPDATE): %s %v, %v %s", isp, path, risMsgUpdate.Announcements, risMsgUpdate.Origin)
				dbObj := TransitRoute{
					Timestamp:     time.Unix(int64(sec), int64(dec*(1e9))),
					TransitAS:     transitAS,
					ISPAS:         path[0],
					DestinationAS: lastPath,
					Json:          string(x),
					Prefixes:      strings.Join(prefixes, ",")}
				db.Create(&dbObj)

				for subscription := range subscriptions {
					subscription <- &dbObj
				}
			default:
				log.Printf("Unhandled bgpmsgtype %s %+v\n", msg.BgpMsgType, msg)
			}
		default:
			log.Printf("Unhandled type %s %+v\n", msg.Type, msg)

		}
	}
}
