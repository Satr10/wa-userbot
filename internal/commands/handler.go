package commands

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Satr10/wa-userbot/internal/config"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// CommandFunc defines the signature for command handler functions.
type CommandFunc func(Command) (whatsmeow.SendResponse, error)

var Commands = registerCommands()

const Footer = "\n\n_pesan otomatis oleh bot_"

// Handler manages command registration and execution.
type Handler struct {
	client   *whatsmeow.Client
	registry map[string]Command
	logger   waLog.Logger
	prefix   string
	cfg      config.Config
	locTime  *time.Location
}

// NewHandler creates a new command handler.
func NewHandler(client *whatsmeow.Client, logger waLog.Logger, config config.Config) (*Handler, error) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Jika lokasi tidak ditemukan, aplikasi tidak bisa berfungsi dengan benar.
		// Sebaiknya hentikan aplikasi di sini.
		return nil, fmt.Errorf("gagal memuat lokasi Asia/Jakarta: %w", err)
	}

	h := &Handler{
		client:   client,
		registry: make(map[string]Command),
		logger:   logger,
		prefix:   ".",
		cfg:      config,
		locTime:  loc,
	}
	// removed moved to commands.go
	// Register commands here
	// h.RegisterCommand("ping", PingCommand)

	return h, nil
}

// RegisterCommand adds a new command to the registry.
// func (h *Handler) RegisterCommand(name string, handlerFunc CommandFunc) {
// 	h.registry[strings.ToLower(name)] = handlerFunc
// 	h.logger.Infof("Registered command: %s%s", h.prefix, name)
// }

func (h *Handler) checkPermission(senderJID types.JID, chatJID types.JID, command *Command) bool {
	userLevel := h.getUserLevel(senderJID, chatJID)
	return userLevel >= int(command.PermissionLevel)
}

func (h *Handler) getUserLevel(senderJID types.JID, chatJID types.JID) int {

	// TODO: finish this later
	if senderJID.String() == fmt.Sprintf("%s@s.whatsapp.net", h.cfg.OwnerID) {
		return int(Owner)
	}

	// TODO: add check for group admin

	return int(Everyone)
}

// HandleEvent processes incoming message events to check for commands.
func (h *Handler) HandleEvent(evt *events.Message) {
	// h.client.MarkRead([]types.MessageID{evt.Info.ID}, time.Now(), evt.Info.Chat, evt.Info.Sender)
	// ignore message from self
	// if evt.Info.IsFromMe {
	// 	return
	// }

	// remove because now we have PermissionLevel
	// ignore if message not from self
	// if !evt.Info.IsFromMe {
	// 	return
	// }

	// mendapatkan pesan dari extended message
	msgText := ""
	if evt.Message.GetConversation() != "" {
		msgText = evt.Message.GetConversation()
	} else if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.Text != nil {
		msgText = evt.Message.ExtendedTextMessage.GetText()
	} else if evt.Message.ImageMessage != nil {
		h.logger.Infof("Image Message Retrieved")
		// h.HandleImage(evt)
	} else {
		// Add support for other message types if needed (e.g., captions)
		return
	}

	// Trim whitespace and check for prefix
	trimmedText := strings.TrimSpace(msgText)
	if strings.HasPrefix(trimmedText, h.prefix) {
		// pass jika ini adalah command ke HandleCommand
		h.HandleCommand(trimmedText, evt)
	}

	// jika bukan command akan di pass ke MessageHandler
	h.MessageHandler(evt)
}

func (h *Handler) HandleCommand(trimmedText string, evt *events.Message) {

	// Split into command and arguments
	parts := strings.Fields(trimmedText)
	if len(parts) == 0 {
		return // Empty command
	}

	// mendapatkan command dan argumen
	commandName := strings.ToLower(strings.TrimPrefix(parts[0], h.prefix))
	args := parts[1:]

	// Look up command in registry
	command, exists := Commands[commandName]
	if !exists {
		// Optionally send "unknown command" message
		h.logger.Infof("Unknown command received: %s", commandName)
		// _, _ = h.client.SendMessage(context.Background(), evt.Info.Chat, &waProto.Message{Conversation: proto.String("Unknown command.")})
		return
	}
	// check command PermissionLevel
	if !h.checkPermission(evt.Info.Sender, evt.Info.Chat, command) {
		return
	}

	h.logger.Infof("Executing command '%s' from %s with args: %v", commandName, evt.Info.Sender, args)

	// Execute command in a goroutine to avoid blocking the event handler
	go func() {
		ctx := context.Background()
		_, err := command.Handler(Command{ctx: ctx, evt: evt, client: h.client, args: args})
		if err != nil {
			h.logger.Errorf("Error executing command '%s': %v", commandName, err)
			// Send error message to user
			SendTextMessage(TextMessage{
				ctx:    ctx,
				client: h.client,
				evt:    evt,
				text:   fmt.Sprintf("err: %v", err),
			})
		}
	}()
}

func (h *Handler) MessageHandler(evt *events.Message) {
	h.AFKHandler(evt)
}

// TODO: need improvements
func (h *Handler) AFKHandler(evt *events.Message) {
	//check if message is in pm or mention in group

	// Check for mentions in group messages
	if evt.Info.IsGroup {
		mentioned := false
		if evt.Message.ExtendedTextMessage != nil {
			mentioned = slices.Contains(evt.Message.ExtendedTextMessage.GetContextInfo().GetMentionedJID(), h.client.Store.LID.String())
		}
		if !mentioned {
			h.logger.Infof("No mention in group message, ignoring")
			return
		}
		if mentioned {
			h.logger.Infof("mentioned in: %s, by %s", evt.Info.Chat, evt.Info.Sender)
		}
	}
	//check time for automated afk message 22.00-07.00
	now := time.Now().In(h.locTime)

	// 2. Buat batas waktu untuk HARI INI secara spesifik
	// Ini membuat objek waktu tanpa harus mem-parsing string, lebih efisien.
	startAFK := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, h.locTime)
	endAFK := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, h.locTime)

	// 3. Lakukan pengecekan yang presisi
	if now.After(startAFK) || now.Before(endAFK) {
		rawText := fmt.Sprintf("Hai! 👋 Terima kasih atas pesannya. Saat ini saya sedang dalam mode istirahat (22.00 - 07.00) dan semua notifikasi sedang nonaktif. Pesan Anda sudah diterima dengan baik dan akan saya balas besok pagi ya. Terima kasih!%s", Footer)
		textMsg := TextMessage{
			ctx:    context.Background(),
			evt:    evt,
			client: h.client,
			text:   rawText,
		}
		_, err := ReplyToTextMesssage(textMsg)
		if err != nil {
			h.logger.Errorf("error sending afk message, err: %s", err)
		}
	}
}
