package commands

import (
	"go.mau.fi/whatsmeow"
)

const Footer = "\n\n_pesan otomatis oleh bot_"

// PingCommand handles the ping command.
// It's now a method of the Handler struct.
func (h *Handler) PingCommand(c Command) (whatsmeow.SendResponse, error) {
	// It can now access dependencies from the handler, for example: h.logger.Infof("Pinging...")
	return ReplyToTextMesssage(TextMessage{ctx: c.ctx, client: c.client, evt: c.evt, text: "Pong" + Footer})
}