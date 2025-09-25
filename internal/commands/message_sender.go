package commands

import (
	"context"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type TextMessage struct {
	ctx    context.Context
	client *whatsmeow.Client
	evt    *events.Message
	text   string
}

func SendTextMessage(t TextMessage) (whatsmeow.SendResponse, error) {
	chatJID := t.evt.Info.Chat
	t.client.SendChatPresence(chatJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	defer t.client.SendChatPresence(chatJID, types.ChatPresencePaused, types.ChatPresenceMediaText)

	msg := &waE2E.Message{Conversation: proto.String(t.text)}
	return t.client.SendMessage(t.ctx, chatJID, msg)
}

func ReplyToTextMesssage(t TextMessage) (whatsmeow.SendResponse, error) {
	recipient := t.evt.Info.Chat
	t.client.SendChatPresence(recipient, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	defer t.client.SendChatPresence(recipient, types.ChatPresencePaused, types.ChatPresenceMediaText)

	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &t.text,
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:      proto.String(t.evt.Info.ID),
				Participant:   proto.String(t.evt.Info.Sender.String()),
				QuotedMessage: t.evt.Message,
			},
		},
	}
	return t.client.SendMessage(t.ctx, recipient, msg)
}

func EditMessage(t TextMessage, originalMsgID types.MessageID) (whatsmeow.SendResponse, error) {
	recipient := t.evt.Info.Chat
	t.client.SendChatPresence(recipient, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	defer t.client.SendChatPresence(recipient, types.ChatPresencePaused, types.ChatPresenceMediaText)

	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &t.text,
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:      proto.String(t.evt.Info.ID),
				Participant:   proto.String(t.evt.Info.Sender.String()),
				QuotedMessage: t.evt.Message,
			},
		},
	}
	edit := t.client.BuildEdit(recipient, originalMsgID, msg)
	return t.client.SendMessage(t.ctx, recipient, edit)
}
