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
	if m.ReplyTo.Text == "Choose your time-to-live:" {
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
	_ = b.Notify(m.Chat, tb.Typing)
	_, _ = b.Send(m.Chat, "You need to promote me to admin first.")
}

func startHandler(m *tb.Message) {
	_ = b.Notify(m.Chat, tb.Typing)
	_, _ = b.Send(m.Chat, "Welcome! I can help deleting group/channel messages. Please add me to your group as admin.")
}

func helpHandler(m *tb.Message) {
	_ = b.Notify(m.Chat, tb.Typing)

	if m.Private() {
		_, _ = b.Send(m.Chat, "Please add me to a group/channel as admin!")
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
		_, _ = b.Send(m.Chat, "Please add me as admin!")
	}
}

func settingsHandler(m *tb.Message) {
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
	var isAdmin = false
	var senderAdmin = false
	if m.Chat.Type == "channel" {
		isAdmin = true
		senderAdmin = true
	} else {
		admins, _ := b.AdminsOf(m.Chat)

		for _, admin := range admins {
			switch admin.User.ID {
			case m.Sender.ID:
				senderAdmin = true
			case b.Me.ID:
				isAdmin = true
			default:
				isAdmin = false
				senderAdmin = false
			}

		}
	}
	//log.Infof("User %d on %s  permission is %v", m.Chat.ID, m.Chat.Type, canSubscribe)

	if !(isAdmin && senderAdmin) {
		// log.Warnf("Denied subscribe request for: %d", m.Sender.ID)
		_ = b.Notify(m.Chat, tb.Typing)
		_, _ = b.Send(m.Chat, "Are you admin? Please promote me as admin.")
		return false
	}
	return true
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
