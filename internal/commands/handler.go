package commands

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/Satr10/wa-userbot/internal/ai"
	aitools "github.com/Satr10/wa-userbot/internal/ai_tools"
	"github.com/Satr10/wa-userbot/internal/config"
	"github.com/Satr10/wa-userbot/internal/permissions"
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
	gemini   *ai.Gemini
	urlRegex *regexp.Regexp
	log      *slog.Logger
	perm     *permissions.Manager
}

// NewHandler creates a new command handler.
func NewHandler(client *whatsmeow.Client, logger waLog.Logger, config config.Config, permManager *permissions.Manager) (*Handler, error) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, fmt.Errorf("gagal memuat lokasi Asia/Jakarta: %w", err)
	}

	aiTools := aitools.NewTools(config.GSBAPIKey)
	newGemini, err := ai.NewGemini(context.TODO(), config.GeminiAPIKey, ai.UrlCheckSystemPrompt, aiTools)
	if err != nil {
		return nil, err
	}

	urlRegex := regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.-]*:(//)?[^\s]*|\b(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}\b(?:/[^\s]*)?`)

	h := &Handler{
		client:   client,
		registry: make(map[string]*Command), // Changed to hold pointers
		logger:   logger,
		prefix:   ".",
		cfg:      config,
		locTime:  loc,
		gemini:   newGemini,
		urlRegex: urlRegex,
		perm:     permManager,
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
	h.registry["edit"] = &Command{
		PermissionLevel: Everyone,
		Handler:         h.EditMsgTest,
	}
	// Perintah manajemen tetap hanya untuk Owner
	h.registry["addgroup"] = &Command{
		PermissionLevel: Owner,
		Handler:         h.AddGroupCommand,
	}
	h.registry["delgroup"] = &Command{
		PermissionLevel: Owner,
		Handler:         h.DelGroupCommand,
	}

	// Register other commands here in the future
	h.logger.Infof("Registered %d commands", len(h.registry))
}

func (h *Handler) checkPermission(senderJID types.JID, chatJID types.JID, command *Command) bool {
	userLevel := h.getUserLevel(senderJID, chatJID)
	if userLevel >= int(command.PermissionLevel) {
		return true
	}

	if command.PermissionLevel == CertainChat {
		// Gunakan manajer izin yang baru
		if h.perm.IsGroupAllowed(chatJID.String()) || h.perm.IsUserAllowed(senderJID.String()) {
			return true
		}
	}

	return false
}

func (h *Handler) getUserLevel(senderJID types.JID, chatJID types.JID) int {
	// Ambil OwnerID dari config
	if senderJID.String() == h.cfg.OwnerID || senderJID.String() == h.cfg.OwnerLID {
		return int(Owner)
	}

	//TODO: TAMBAHKAN UNTUK CEK APAKAH USER ADALAH ADMIN GRUP

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

	h.MessageHandler(evt, msgText)
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
	fmt.Println(evt.Info.Sender.ToNonAD())

	if !h.checkPermission(evt.Info.Sender.ToNonAD(), evt.Info.Chat, command) {
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

func (h *Handler) MessageHandler(evt *events.Message, msgText string) {
	h.AFKHandler(evt)
	h.UrlScan(evt, msgText)
}

// AFKHandler menangani logika untuk membalas pesan secara otomatis saat di luar jam kerja.
func (h *Handler) AFKHandler(evt *events.Message) {
	// Guard Clause: Abaikan pesan dari diri sendiri.
	if evt.Info.IsFromMe {
		return
	}

	// Periksa apakah saat ini dalam periode AFK.
	now := time.Now().In(h.locTime)
	if !h.isWithinAFKHours(now) {
		return
	}

	// Tentukan apakah bot harus membalas:
	// - Selalu balas di Direct Message (DM).
	// - Di grup, hanya balas jika di-mention.
	isDM := !evt.Info.IsGroup
	isMentioned := evt.Info.IsGroup && h.isBotMentioned(evt)

	if !isDM && !isMentioned {
		return
	}

	// Jika semua kondisi terpenuhi, kirim pesan AFK.
	h.sendAFKMessage(evt)
}

// isBotMentioned memeriksa apakah JID bot ada dalam daftar mention pesan.
func (h *Handler) isBotMentioned(evt *events.Message) bool {
	extMsg := evt.Message.ExtendedTextMessage
	if extMsg == nil || extMsg.ContextInfo == nil {
		return false
	}

	mentionedJIDs := extMsg.GetContextInfo().GetMentionedJID()
	if len(mentionedJIDs) == 0 {
		return false
	}

	// Ambil JID bot sekali saja untuk perbandingan.
	myJID := h.client.Store.GetJID().ToNonAD().String()
	myLID := h.client.Store.LID.String()

	// Gunakan perulangan untuk efisiensi jika daftar mention besar.
	// slices.Contains melakukan hal serupa di baliknya.
	for _, mentioned := range mentionedJIDs {
		if mentioned == myJID || mentioned == myLID {
			return true
		}
	}

	return false
}

// isWithinAFKHours memeriksa apakah waktu yang diberikan berada dalam rentang AFK (22:00 - 07:00).
func (h *Handler) isWithinAFKHours(t time.Time) bool {
	// Logika ini menangani kasus semalam (misal: 22:00 - 07:00 keesokan harinya)
	hour := t.Hour()
	return hour >= 22 || hour < 7
}

// sendAFKMessage merakit dan mengirimkan pesan balasan AFK.
func (h *Handler) sendAFKMessage(evt *events.Message) {
	rawText := fmt.Sprintf("Hai! 👋 Terima kasih atas pesannya. Saat ini saya sedang dalam mode istirahat (22.00 - 07.00) dan semua notifikasi sedang nonaktif. Pesan Anda sudah diterima dengan baik dan akan saya balas besok pagi ya. Terima kasih!%s", Footer)

	textMsg := TextMessage{
		ctx:    context.Background(),
		evt:    evt,
		client: h.client,
		text:   rawText,
	}

	if _, err := ReplyToTextMesssage(textMsg); err != nil {
		h.logger.Errorf("error sending afk message, err: %s", err)
	}
}

func (h *Handler) UrlScan(evt *events.Message, msgText string) {
	if evt.Info.IsFromMe || h.perm.IsGroupAllowed(evt.Info.Chat.ToNonAD().String()) {
		allUrls := h.urlRegex.FindAll([]byte(msgText), -1)
		for _, url := range allUrls {
			hash := sha256.Sum256([]byte(url))
			id := hex.EncodeToString(hash[:])

			var initialPrompt string
			if len(url) > 512 {
				initialPrompt = fmt.Sprintf("Mulai investigasi untuk URL dengan ID: %s. URL SANGAT PANJANG (%d karakter), indikasikan potensi serangan DoS atau upaya mengaburkan URL asli. URL: saya tidak masukan karena berbahaya bisa membuat program crash atau error.JIKA PANJANG KARAKTER URL KETERLALUAN MENURUTMU, BERI PERINGATAN KERAS!!! KEPADA USER JAHIL KARENA INI DAPAT MENYEBABKAN PROGRAM CRASH", id, len(url))
			} else {
				initialPrompt = fmt.Sprintf("Mulai investigasi untuk URL: %s dengan ID: %s", url, id)
			}

			result, err := h.gemini.URLScan(context.TODO(), initialPrompt, id)
			if err != nil {

				errorMessage := TextMessage{
					ctx:    context.TODO(),
					client: h.client,
					evt:    evt,
					text:   "error scanning url" + Footer,
				}
				ReplyToTextMesssage(errorMessage)

				fmt.Printf("error scanning url %s: %v\n", url, err)
				continue // Skip to the next URL
			}
			textMessage := TextMessage{
				ctx:    context.TODO(),
				client: h.client,
				evt:    evt,
				text:   result.FormatWhatsAppMessage() + Footer,
			}
			ReplyToTextMesssage(textMessage)
		}
	}
}
