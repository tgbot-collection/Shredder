// SensitiveCleaner - handler
// 2021-02-04 13:43
// Benny <benny.think@gmail.com>

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"time"
)

var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	Addr:     fmt.Sprintf("%s:6379", redisHost),
	Password: "", // no password set
	DB:       1,  // use default DB
})

func messageHandler(m *tb.Message) {
	if m.ReplyTo != nil && m.ReplyTo.Text == "Choose your time-to-live:" {
		if !permissionCheck(m) {
			return
		}
		data := readJSON()
		re := regexp.MustCompile(`\d+`)
		result := re.FindAllString(m.Text, -1)[0]
		i, _ := strconv.Atoi(result)
		strID := strconv.FormatInt(m.Chat.ID, 10)
		data[strID] = int64(i * 3600)
		saveJSON(data)
		_, _ = b.Send(m.Chat, "Oorah!", &tb.ReplyMarkup{ReplyKeyboardRemove: true})

		return
	}
	log.Infof("Receiving message ID %d at %s", m.ID, time.Now())
	// check group
	if m.Private() {
		_, _ = b.Send(m.Chat, "Add me to a group or channel, and I'll start working.")
		return
	}
	rdb.HSet(ctx, fmt.Sprintf("%v", m.Chat.ID), m.ID, time.Now().Unix())
}

func onJoinHandler(m *tb.Message) {
	rdb.HSet(ctx, fmt.Sprintf("%v", m.Chat.ID), m.ID, time.Now().Unix())
	_ = b.Notify(m.Chat, tb.Typing)
	_, _ = b.Send(m.Chat, "You need to promote me to admin first. And then optionally call /settings to set TTL."+
		"Default TTL is 48h.")
}

func startHandler(m *tb.Message) {
	rdb.HSet(ctx, fmt.Sprintf("%v", m.Chat.ID), m.ID, time.Now().Unix())
	_ = b.Notify(m.Chat, tb.Typing)
	_, _ = b.Send(m.Chat, "Welcome! I can help deleting group/channel messages. Please add me to your group as admin."+
		"See /settings for more.")
}

func helpHandler(m *tb.Message) {
	rdb.HSet(ctx, fmt.Sprintf("%v", m.Chat.ID), m.ID, time.Now().Unix())
	_ = b.Notify(m.Chat, tb.Typing)

	helpMsg := `A bot that will help you automatically delete group messages.
			You need to:
			1. add me to a group/channel as admin.
			2. optionally set TTL, default is 48h`
	if m.Private() {
		_, _ = b.Send(m.Chat, helpMsg)
		return
	}

	data := readJSON()
	strID := strconv.FormatInt(m.Chat.ID, 10)
	value := data[strID] / 3600

	var usable = false
	admins, _ := b.AdminsOf(m.Chat)
	for _, admin := range admins {
		if admin.User.ID == b.Me.ID {
			usable = true
		}
	}
	if usable || m.Chat.Type == "channel" {
		_, _ = b.Send(m.Chat, fmt.Sprintf("I'm working, your TTL is %dh", value))
	} else {
		_, _ = b.Send(m.Chat, helpMsg)
	}
}

func settingsHandler(m *tb.Message) {
	rdb.HSet(ctx, fmt.Sprintf("%v", m.Chat.ID), m.ID, time.Now().Unix())
	if !permissionCheck(m) {
		return
	}

	var menu = &tb.ReplyMarkup{ForceReply: true, ResizeReplyKeyboard: true}

	var rows []tb.Row
	var btns []tb.Btn
	var count = 1

	for i := 1; i <= 48; i += 2 {
		// 1 3 5 7
		btn := menu.Text(fmt.Sprintf("%dh", i))
		btns = append(btns, btn)
		if count > 5 {
			rows = append(rows, menu.Row(btns...))
			btns = []tb.Btn{}
			count = 1
		} else {
			count += 1
		}
	}
	menu.Reply(rows...)
	_, _ = b.Send(m.Chat, "Choose your time-to-live:", menu)
}

func permissionCheck(m *tb.Message) bool {
	var botAdmin = false
	var senderAdmin = false
	if m.Chat.Type == "channel" {
		botAdmin = true
		senderAdmin = true
	} else {
		admins, _ := b.AdminsOf(m.Chat)
		var adminMap = make(map[int]int)
		for _, admin := range admins {
			adminMap[admin.User.ID] = 1
		}

		if _, ok := adminMap[m.Sender.ID]; ok {
			senderAdmin = true
		}
		if _, ok := adminMap[b.Me.ID]; ok {
			botAdmin = true
		}

	}

	_ = b.Notify(m.Chat, tb.Typing)

	if !botAdmin {
		_, _ = b.Send(m.Chat, "Please promote me to admin.")
		return false
	} else if !senderAdmin {
		_, _ = b.Send(m.Chat, "Are you admin? ")
		return false
	} else {
		return true
	}

}

func readJSON() userConfig {
	log.Debugf("Read json file...")
	jsonFile, _ := os.Open("settings.json")
	decoder := json.NewDecoder(jsonFile)

	var config = make(userConfig)
	err = decoder.Decode(&config)
	_ = jsonFile.Close()
	return config
}

func saveJSON(current userConfig) {
	e, _ := json.MarshalIndent(current, "", "\t")
	err := ioutil.WriteFile("settings.json", e, 0644)
	if err != nil {
		log.Errorf("Write json failed %v", err)
	}
}
