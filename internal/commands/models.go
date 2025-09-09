package commands

import (
	"context"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type Command struct {
	ctx    context.Context
	client *whatsmeow.Client
	evt    *events.Message
	args   []string
}
