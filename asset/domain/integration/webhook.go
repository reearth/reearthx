package integration

import (
	"net/url"
	"time"

	"github.com/reearth/reearthx/asset/domain/event"
)

type Webhook struct {
	updatedAt time.Time
	url       *url.URL
	trigger   WebhookTrigger
	name      string
	secret    string
	id        WebhookID
	active    bool
}

type WebhookTrigger map[event.Type]bool

func (w *Webhook) ID() WebhookID {
	return w.id
}

func (w *Webhook) Name() string {
	return w.name
}

func (w *Webhook) SetName(name string) {
	w.name = name
}

func (w *Webhook) URL() *url.URL {
	return w.url
}

func (w *Webhook) SetURL(url *url.URL) {
	w.url = url
}

func (w *Webhook) Active() bool {
	return w.active
}

func (w *Webhook) SetActive(active bool) {
	w.active = active
}

func (w *Webhook) Trigger() WebhookTrigger {
	return w.trigger
}

func (w *Webhook) SetTrigger(trigger WebhookTrigger) {
	w.trigger = trigger
}

func (w *Webhook) UpdatedAt() time.Time {
	if w.updatedAt.IsZero() {
		return w.id.Timestamp()
	}
	return w.updatedAt
}

func (w *Webhook) CreatedAt() time.Time {
	return w.id.Timestamp()
}

func (w *Webhook) SetUpdatedAt(updatedAt time.Time) {
	w.updatedAt = updatedAt
}

func (w *Webhook) Secret() string {
	return w.secret
}

func (w *Webhook) SetSecret(secret string) {
	w.secret = secret
}

func (w *Webhook) Clone() *Webhook {
	if w == nil {
		return nil
	}

	var u *url.URL = nil
	if w.url != nil {
		u, _ = url.Parse(w.url.String())
	}
	return &Webhook{
		id:        w.id.Clone(),
		name:      w.name,
		url:       u,
		active:    w.active,
		trigger:   w.trigger,
		updatedAt: w.updatedAt,
		secret:    w.secret,
	}
}

func (t WebhookTrigger) IsActive(et event.Type) bool {
	return t[et]
}

func (t WebhookTrigger) Enable(et event.Type) {
	t[et] = true
}

func (t WebhookTrigger) Disable(et event.Type) {
	delete(t, et)
}
