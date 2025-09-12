package bot

import (
	"go.mau.fi/whatsmeow/types/events"
)

func (b *Bot) eventHandler(evt any) {
	switch v := evt.(type) {
	case *events.Connected:
		b.logger.Infof("Device connected")

	case *events.Disconnected:
		b.logger.Warnf("Device disconnected")

	case *events.Message:
		// logger.Info("Message received", "message", v.Message)
		b.cmdHandler.HandleEvent(v)

	case *events.Receipt:

	case *events.Presence:
		b.logger.Infof("Presence")

	case *events.LoggedOut:
		b.logger.Warnf("Logged out, reason: %v", v.Reason)

	default:
		b.logger.Debugf("Ignored debug, event: %v", evt)
	}
}
