// SensitiveCleaner - scheduler
// 2021-02-04 13:48
// Benny <benny.think@gmail.com>

package main

import (
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
	"math/rand"
	"strconv"
	"time"
)

func scheduler() {
	keys := rdb.Keys(ctx, "*")
	groupID, _ := keys.Result()
	log.Infoln("Running delete scheduler")
	for _, g := range groupID {
		message, _ := rdb.HGetAll(ctx, g).Result()
		for mid, ts := range message {
			intts, _ := strconv.Atoi(ts)
			intcid, _ := strconv.Atoi(g)
			data := readJSON()
			v := data[g]
			if time.Now().Unix()-int64(intts) > v {
				log.Debugln("Deleting...")
				var msg tb.StoredMessage
				msg.ChatID = int64(intcid)
				msg.MessageID = mid
				rand.Seed(time.Now().UnixNano())
				time.Sleep(time.Second * time.Duration(rand.Intn(2)))
				_ = b.Delete(msg)
				// unset redis
				rdb.HDel(ctx, g, mid)

			}
		}
	}

}
