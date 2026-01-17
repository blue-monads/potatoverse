package stdio

import (
	"log/slog"

	"github.com/blue-monads/potatoverse/backend/services/mailer"
)

type Mailer struct {
	slog *slog.Logger
}

func NewMailer(slog *slog.Logger) *Mailer {
	return &Mailer{
		slog: slog,
	}
}

func (m *Mailer) Send(to string, subject string, body mailer.MessageBody) error {
	text, err := body.AsText()
	if err != nil {
		return err
	}

	m.slog.Info("@Send", "to", to, "subject", subject, "body", text)

	return nil
}
