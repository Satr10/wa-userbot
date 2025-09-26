package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Satr10/wa-userbot/internal/bot"
	"github.com/Satr10/wa-userbot/internal/config"
	"github.com/Satr10/wa-userbot/internal/permissions"
	"github.com/Satr10/wa-userbot/internal/server"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	logger := waLog.Stdout("Main", "Info", true)
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		logger.Errorf("error no .env file(ignore if you use real enviroment variables)")
	}

	logger.Infof("Using database URI: %s", os.Getenv("POSTGRES_URI"))

	permManager, err := permissions.NewManager("/tmp/permissions.json")
	if err != nil {
		logger.Errorf("error creating new permissions manager err: %v", err)
	}
	botInstance, err := bot.NewBot(logger, cfg, permManager)
	if err != nil {
		logger.Errorf("Error creating new bot instance, err: %v", err)
		return
	}

	go server.StartServer("8080")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Infof("Shutting down...")
	botInstance.Disconnect()
	server.Shutdown(context.Background())
}
