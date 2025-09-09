package commands

import (
	"go.mau.fi/whatsmeow"
)

func PingCommand(c Command) (whatsmeow.SendResponse, error) {
	return ReplyToTextMesssage(TextMessage{ctx: c.ctx, client: c.client, evt: c.evt, text: "Pong"})
}
