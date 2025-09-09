package bot

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Satr10/wa-userbot/internal/commands"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Bot struct {
	client     *whatsmeow.Client
	logger     waLog.Logger
	cmdHandler *commands.Handler
	botUptime  time.Time
}

func NewBot(logger waLog.Logger) (newBot *Bot, err error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:database.db?_foreign_keys=on", dbLog)
	if err != nil {
		return nil, err
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, err
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	if client.Store.ID == nil {
		// jika tidak ada id login dari awal
		qrChan, _ := client.GetQRChannel(ctx)
		err = client.Connect()
		if err != nil {
			return nil, err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal

			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			return nil, err
		}
	}
	botInstance := &Bot{
		logger:     logger,
		client:     client,
		cmdHandler: commands.NewHandler(client, logger),
		botUptime:  time.Now(),
	}
	client.SendPresence(types.PresenceAvailable)
	client.AddEventHandler(botInstance.eventHandler)
	return botInstance, nil
}

func (b *Bot) Client() *whatsmeow.Client {
	return b.client
}

// disconnect untuk graceful shutdown
func (b *Bot) Disconnect() {
	b.logger.Infof("Disconnecting...")
	b.client.Disconnect()

}
