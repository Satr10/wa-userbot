package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Satr10/wa-userbot/internal/bot"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	logger := waLog.Stdout("Main", "Info", true)

	botInstance, err := bot.NewBot(logger)
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
