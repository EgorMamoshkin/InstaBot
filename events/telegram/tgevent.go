package telegram

import (
	"InstaBot/clients/tgclient"
	"InstaBot/events"
	"InstaBot/lib/er"
	"InstaBot/storage"
	"context"
	"errors"
)

type Processor struct {
	tg      *tgclient.Client
	offset  int
	storage storage.Storage
}

type Meta struct {
	ChatID   int
	Username string
}

var ErrUnknownEventType = errors.New("unknown event type")
var ErrUnknownMetaType = errors.New("unknown meta type")

func New(tgClient *tgclient.Client, storage storage.Storage) *Processor {
	return &Processor{
		tg:      tgClient,
		storage: storage,
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, er.Wrap("can't get updates:", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	for _, upd := range updates {
		res = append(res, event(upd))
	}

	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}

func (p *Processor) Process(ctx context.Context, event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(ctx, event)
	default:
		return er.Wrap("can't handle event", ErrUnknownEventType)
	}
}

func (p *Processor) processMessage(ctx context.Context, event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return er.Wrap("can't handle message:", err)
	}

	if err := p.execCmd(ctx, event.Text, meta.ChatID, meta.Username); err != nil {
		return er.Wrap("can't process the message:", err)
	}

	return nil
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, er.Wrap("can't get meta:", ErrUnknownMetaType)
	}

	return res, nil
}

func event(upd tgclient.Update) events.Event {
	updType := FetchType(upd)
	res := events.Event{
		Type: updType,
		Text: FetchText(upd),
	}

	if updType == events.Message {
		res.Meta = Meta{
			ChatID:   upd.Message.Chat.ID,
			Username: upd.Message.From.UserName,
		}
	}

	return res
}

func FetchText(upd tgclient.Update) string {
	if upd.Message == nil {
		return ""
	}

	return upd.Message.Text
}

func FetchType(upd tgclient.Update) events.Type {
	if upd.Message == nil {
		return events.Unknown
	}

	return events.Message
}
