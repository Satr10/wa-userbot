package commands

import (
	"go.mau.fi/whatsmeow"
)

func registerCommands() map[string]*Command {
	commands := make(map[string]*Command)

	commands["ping"] = &Command{
		PermissionLevel: Everyone,
		Handler:         PingCommand,
	}

	return commands
}

func PingCommand(c Command) (whatsmeow.SendResponse, error) {
	return ReplyToTextMesssage(TextMessage{ctx: c.ctx, client: c.client, evt: c.evt, text: "Pong"})
}
