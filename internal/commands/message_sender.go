package commands

import (
	"context"

	"github.com/davidbyttow/govips/v2/vips"
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
type ImageMessage struct {
	ctx        context.Context
	client     *whatsmeow.Client
	evt        *events.Message
	caption    string
	imageBytes []byte
	mimeType   string
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

func SendImage(i ImageMessage) (whatsmeow.SendResponse, error) {
	recipient := i.evt.Info.Chat
	i.client.SendChatPresence(recipient, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	defer i.client.SendChatPresence(recipient, types.ChatPresencePaused, types.ChatPresenceMediaText)

	resp, err := i.client.Upload(context.Background(), i.imageBytes, whatsmeow.MediaImage)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}

	imgMsg := &waE2E.ImageMessage{
		Caption:  proto.String(i.caption),
		Mimetype: proto.String(i.mimeType),
		// JPEGThumbnail: test,
		URL:           &resp.URL,
		DirectPath:    &resp.DirectPath,
		MediaKey:      resp.MediaKey,
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    &resp.FileLength,
	}

	return i.client.SendMessage(i.ctx, recipient, &waE2E.Message{ImageMessage: imgMsg})
}

// generateThumbnailVips creates a JPEG thumbnail using the govips library.
func generateThumbnailVips(imageBytes []byte) ([]byte, error) {
	// It's recommended to call Startup once per application start.
	// If you call this function frequently, move Startup/Shutdown
	// to your application's main function.
	vips.Startup(nil)
	defer vips.Shutdown()

	// 1. Load the image from the byte buffer
	img, err := vips.NewImageFromBuffer(imageBytes)
	if err != nil {
		return nil, err
	}
	defer img.Close() // Make sure to close the image to free C memory

	// 2. Create a thumbnail, preserving aspect ratio.
	// The third argument is the interpolator, Lanczos3 is high quality.
	err = img.Thumbnail(160, 160, vips.Interesting(vips.KernelLanczos3))
	if err != nil {
		return nil, err
	}

	// 3. Export the thumbnail to a new buffer as a JPEG
	// We use default JPEG export parameters.
	params := vips.NewJpegExportParams()
	thumbnailBytes, _, err := img.ExportJpeg(params)
	if err != nil {
		return nil, err
	}

	return thumbnailBytes, nil
}
