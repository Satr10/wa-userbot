package commands

import (
	"fmt"
	"go.mau.fi/whatsmeow"
)

const Footer = "\n\n_pesan otomatis oleh bot_"

// Helper function untuk mengirim pesan reply dengan format yang konsisten
func (h *Handler) sendReply(c Command, message string) (whatsmeow.SendResponse, error) {
	textMsg := TextMessage{
		ctx:    c.ctx,
		client: c.client,
		evt:    c.evt,
		text:   message + Footer,
	}
	return ReplyToTextMesssage(textMsg)
}

// Helper function untuk edit pesan dengan format yang konsisten
func (h *Handler) editMessage(c Command, message string, messageID string) (whatsmeow.SendResponse, error) {
	textMsg := TextMessage{
		ctx:    c.ctx,
		client: c.client,
		evt:    c.evt,
		text:   message + Footer,
	}
	return EditMessage(textMsg, messageID)
}

// PingCommand handles the ping command
func (h *Handler) PingCommand(c Command) (whatsmeow.SendResponse, error) {
	// Log aktivitas jika diperlukan
	// h.logger.Infof("Processing ping command from user: %s", c.evt.Info.Sender.String())

	return h.sendReply(c, "Pong")
}

// EditMsgTest demonstrates message editing functionality
func (h *Handler) EditMsgTest(c Command) (whatsmeow.SendResponse, error) {
	// Kirim pesan pertama
	firstMsg, err := h.sendReply(c, "Pong")
	if err != nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("failed to send initial message: %w", err)
	}

	// Edit pesan dengan konten baru
	return h.editMessage(c, "Pong Edit", firstMsg.ID)
}

// AddUserCommand menambahkan user ke daftar yang diizinkan
// func (h *Handler) AddUserCommand(c Command) (whatsmeow.SendResponse, error) {
// 	userIDs := c.evt.Message.ExtendedTextMessage.GetContextInfo().GetMentionedJID()
// 	// Validasi input
// 	if len(userIDs) == 0 {
// 		return h.sendReply(c, "Penggunaan: .adduser <user_id>")
// 	}
//
// 	// Eksekusi penambahan user
// 	for _, userID := range userIDs {
// 		if err := h.perm.AddAllowedUser(userID); err != nil {
// 			errorMsg := fmt.Sprintf("Gagal menambahkan pengguna: %v", err)
// 			return h.sendReply(c, errorMsg)
// 		}
// 	}
//
// 	// Kirim pesan sukses
// 	successMsg := fmt.Sprintf("Pengguna %s berhasil ditambahkan.", userIDs)
// 	return h.sendReply(c, successMsg)
// }
//
// // DelUserCommand menghapus user dari daftar yang diizinkan
// func (h *Handler) DelUserCommand(c Command) (whatsmeow.SendResponse, error) {
// 	userIDs := c.evt.Message.ExtendedTextMessage.GetContextInfo().GetMentionedJID()
// 	// Validasi input
// 	if len(userIDs) == 0 {
// 		return h.sendReply(c, "Penggunaan: .deluser <user_id>")
// 	}
//
// 	// Eksekusi penghapusan user
// 	for _, userID := range userIDs {
// 		if err := h.perm.RemoveAllowedUser(userID); err != nil {
// 			errorMsg := fmt.Sprintf("Gagal menghapus pengguna: %v", err)
// 			return h.sendReply(c, errorMsg)
// 		}
//
// 	}
// 	// Kirim pesan sukses
// 	successMsg := fmt.Sprintf("Pengguna %s berhasil dihapus.", userIDs)
// 	return h.sendReply(c, successMsg)
// }

// AddGroupCommand menambahkan grup ke daftar yang diizinkan
func (h *Handler) AddGroupCommand(c Command) (whatsmeow.SendResponse, error) {
	// Dapatkan group ID dari context
	groupID := c.evt.Info.Chat.String()

	// Eksekusi penambahan grup
	if err := h.perm.AddAllowedGroup(groupID); err != nil {
		errorMsg := fmt.Sprintf("Gagal menambahkan grup: %v", err)
		return h.sendReply(c, errorMsg)
	}

	// Kirim pesan sukses
	successMsg := fmt.Sprintf("Grup ini (%s) berhasil ditambahkan.", groupID)
	return h.sendReply(c, successMsg)
}

// DelGroupCommand menghapus grup dari daftar yang diizinkan
func (h *Handler) DelGroupCommand(c Command) (whatsmeow.SendResponse, error) {
	// Dapatkan group ID dari context
	groupID := c.evt.Info.Chat.String()

	// Eksekusi penghapusan grup
	if err := h.perm.RemoveAllowedGroup(groupID); err != nil {
		errorMsg := fmt.Sprintf("Gagal menghapus grup: %v", err)
		return h.sendReply(c, errorMsg)
	}

	// Kirim pesan sukses
	successMsg := fmt.Sprintf("Grup ini (%s) berhasil dihapus.", groupID)
	return h.sendReply(c, successMsg)
}

