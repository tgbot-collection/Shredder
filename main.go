package main

import (
	"fmt"
)

import (
	"os"
	"runtime"
	"time"
)

import (
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

var b, err = tb.NewBot(tb.Settings{
	Token:  Token,
	Poller: &tb.LongPoller{Timeout: 10 * time.Second},
})

func main() {

	if err != nil {
		log.Panicf("Please check your network or TOKEN! %v", err)
	}
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)
	Formatter := &log.TextFormatter{
		EnvironmentOverrideColors: true,
		FullTimestamp:             true,
		TimestampFormat:           "2006-01-02 15:04:05",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return fmt.Sprintf("[%s()]", f.Function), ""
		},
	}
	log.SetFormatter(Formatter)
	c := cron.New()
	_, _ = c.AddFunc("*/15 * * * *", scheduler)
	c.Start()
	//  toilet  KeepMe.Run -f smblock
	banner := fmt.Sprintf(`
Sensitive Cleaner
By %s 

Across the Great Wall, we can reach every corner in the world.
`, "BennyThink")
	fmt.Printf("\n %c[1;32m%s%c[0m\n\n", 0x1B, banner, 0x1B)

	b.Handle("/start", startHandler)
	b.Handle("/help", helpHandler)
	b.Handle("/settings", settingsHandler)
	b.Handle(tb.OnText, messageHandler)
	b.Handle(tb.OnChannelPost, messageHandler)
	b.Handle(tb.OnDocument, messageHandler)
	b.Handle(tb.OnAnimation, messageHandler)
	b.Handle(tb.OnAudio, messageHandler)
	b.Handle(tb.OnVideo, messageHandler)
	b.Handle(tb.OnVoice, messageHandler)
	b.Handle(tb.OnVideoNote, messageHandler)
	b.Handle(tb.OnLocation, messageHandler)
	b.Handle(tb.OnSticker, messageHandler)
	b.Handle(tb.OnAddedToGroup, onJoinHandler)

	log.Infoln("I'm running...")

	b.Start()

}
