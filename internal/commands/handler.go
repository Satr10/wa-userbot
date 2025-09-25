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

// Handler manages command registration and execution.
type Handler struct {
	client   *whatsmeow.Client
	registry map[string]*Command // Changed to hold pointers
	logger   waLog.Logger
	prefix   string
	cfg      config.Config
	locTime  *time.Location
}

// NewHandler creates a new command handler.
func NewHandler(client *whatsmeow.Client, logger waLog.Logger, config config.Config) (*Handler, error) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, fmt.Errorf("gagal memuat lokasi Asia/Jakarta: %w", err)
	}

	h := &Handler{
		client:   client,
		registry: make(map[string]*Command), // Changed to hold pointers
		logger:   logger,
		prefix:   ".",
		cfg:      config,
		locTime:  loc,
	}

	h.registerCommands()

	return h, nil
}

// registerCommands initializes and registers all commands.
func (h *Handler) registerCommands() {
	h.registry["ping"] = &Command{
		PermissionLevel: Everyone,
		Handler:         h.PingCommand,
	}

	// Register other commands here in the future
	h.logger.Infof("Registered %d commands", len(h.registry))
}

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
	msgText := ""
	if evt.Message.GetConversation() != "" {
		msgText = evt.Message.GetConversation()
	} else if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.Text != nil {
		msgText = evt.Message.ExtendedTextMessage.GetText()
	} else if evt.Message.ImageMessage != nil {
		h.logger.Infof("Image Message Retrieved")
	} else {
		return
	}

	trimmedText := strings.TrimSpace(msgText)
	if strings.HasPrefix(trimmedText, h.prefix) {
		h.HandleCommand(trimmedText, evt)
		return // It's a command, so we stop further processing
	}

	h.MessageHandler(evt)
}

func (h *Handler) HandleCommand(trimmedText string, evt *events.Message) {
	parts := strings.Fields(trimmedText)
	if len(parts) == 0 {
		return
	}

	commandName := strings.ToLower(strings.TrimPrefix(parts[0], h.prefix))
	args := parts[1:]

	command, exists := h.registry[commandName] // Use the handler's registry
	if !exists {
		h.logger.Infof("Unknown command received: %s", commandName)
		return
	}

	if !h.checkPermission(evt.Info.Sender, evt.Info.Chat, command) {
		return
	}

	h.logger.Infof("Executing command '%s' from %s with args: %v", commandName, evt.Info.Sender, args)

	go func() {
		ctx := context.Background()
		_, err := command.Handler(Command{ctx: ctx, evt: evt, client: h.client, args: args})
		if err != nil {
			h.logger.Errorf("Error executing command '%s': %v", commandName, err)
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

func (h *Handler) AFKHandler(evt *events.Message) {
	if evt.Info.IsGroup {
		mentioned := false
		if evt.Message.ExtendedTextMessage != nil {
			mentioned = slices.Contains(evt.Message.ExtendedTextMessage.GetContextInfo().GetMentionedJID(), h.client.Store.LID.String())
		}
		if !mentioned {
			return
		}
		if mentioned {
			h.logger.Infof("mentioned in: %s, by %s", evt.Info.Chat, evt.Info.Sender)
		}
	}

	now := time.Now().In(h.locTime)
	startAFK := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, h.locTime)
	endAFK := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, h.locTime)

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