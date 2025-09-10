package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Satr10/wa-userbot/internal/bot"
	"github.com/Satr10/wa-userbot/internal/config"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	logger := waLog.Stdout("Main", "Info", true)
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		panic(err)
	}

	botInstance, err := bot.NewBot(logger, cfg)
	if err != nil {
		logger.Errorf("Error creating new bot instance, err: %v", err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Infof("Shutting down...")
	botInstance.Disconnect()
}
