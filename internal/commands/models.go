package commands

import (
	"context"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type PermissionLevel int

const (
	Everyone PermissionLevel = iota
	GroupAdmin
	SuperAdmin
	Owner
)

type Command struct {
	ctx    context.Context
	client *whatsmeow.Client
	evt    *events.Message
	args   []string
	PermissionLevel PermissionLevel
}
