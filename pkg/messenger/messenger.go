package messenger

import (
	"github.com/streadway/amqp"
)

type Message interface {
}

type Messenger struct {
	conn *amqp.Connection
}

func NewMessenger(dsn string) (*Messenger, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, err
	}
	return &Messenger{conn: conn}, nil
}

func (m *Messenger) Publish(queue string, body []byte) error {
	ch, err := m.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		queue,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)
	return err
}
