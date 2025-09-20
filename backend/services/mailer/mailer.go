package mailer

type MessageBody interface {
	AsHTML() (string, error)
	AsText() (string, error)
}

type Mailer interface {
	Send(to string, subject string, body MessageBody) error
}

type SimpleMessage struct {
	HTML string
	Text string
}

func (t *SimpleMessage) AsHTML() (string, error) {
	return t.HTML, nil
}

func (t *SimpleMessage) AsText() (string, error) {
	return t.Text, nil
}
