package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	registry map[string]CommandFunc
	logger   waLog.Logger
	prefix   string
}

// NewHandler creates a new command handler.
func NewHandler(client *whatsmeow.Client, logger waLog.Logger) *Handler {
	h := &Handler{
		client:   client,
		registry: make(map[string]CommandFunc),
		logger:   logger,
		prefix:   ".",
	}
	// Register commands here
	h.RegisterCommand("ping", PingCommand)

	return h
}

// RegisterCommand adds a new command to the registry.
func (h *Handler) RegisterCommand(name string, handlerFunc CommandFunc) {
	h.registry[strings.ToLower(name)] = handlerFunc
	h.logger.Infof("Registered command: %s%s", h.prefix, name)
}

// HandleEvent processes incoming message events to check for commands.
func (h *Handler) HandleEvent(evt *events.Message) {
	h.client.MarkRead([]types.MessageID{evt.Info.ID}, time.Now(), evt.Info.Chat, evt.Info.Sender)
	// ignore message from self
	// if evt.Info.IsFromMe {
	// 	return
	// }

	// ignore if message not from self
	if !evt.Info.IsFromMe {
		return
	}

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
	if !strings.HasPrefix(trimmedText, h.prefix) {
		// handle all message
		// HandleSemuaPesan(context.Background(), h.client, evt, msgText)
		// mungkin return aja ya, biar kode yang dibawah gak dieksekusi(boros resource wkwkwk)
		return // Not a command
	}

	// Split into command and arguments
	parts := strings.Fields(trimmedText)
	if len(parts) == 0 {
		return // Empty command
	}

	// mendapatkan command dan argumen
	commandName := strings.ToLower(strings.TrimPrefix(parts[0], h.prefix))
	args := parts[1:]

	// Look up command in registry
	cmdFunc, exists := h.registry[commandName]
	if !exists {
		// Optionally send "unknown command" message
		h.logger.Infof("Unknown command received: %s", commandName)
		// _, _ = h.client.SendMessage(context.Background(), evt.Info.Chat, &waProto.Message{Conversation: proto.String("Unknown command.")})
		return
	}

	h.logger.Infof("Executing command '%s' from %s with args: %v", commandName, evt.Info.Sender, args)

	// Execute command in a goroutine to avoid blocking the event handler
	go func() {
		ctx := context.Background()
		_, err := cmdFunc(Command{ctx: ctx, evt: evt, client: h.client, args: args})
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
