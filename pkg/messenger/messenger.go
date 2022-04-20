package messenger

import (
	"github.com/streadway/amqp"
)

const (
	AppDeployQueue = "application.deploy"
)

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

// GetMessages return messages from broker and a channel
func (m *Messenger) GetMessages(queue string) (<-chan amqp.Delivery, *amqp.Channel, error) {
	ch, err := m.conn.Channel()
	if err != nil {
		return nil, nil, err
	}
	q, err := ch.QueueDeclare(
		queue,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return nil, nil, err
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		_ = ch.Close()
		return nil, nil, err
	}
	return msgs, ch, nil
}

// CountMessages get the total messages in a queue
func (m *Messenger) CountMessages(queue string) (int, error) {
	ch, err := m.conn.Channel()
	if err != nil {
		return 0, err
	}
	defer ch.Close()
	q, err := ch.QueueInspect(queue)
	if err != nil {
		return 0, err
	}
	return q.Messages, nil
}

// PurgeQueue purge non delivered messages in a queue
func (m *Messenger) PurgeQueue(queue string) error {
	ch, err := m.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	_, err = ch.QueuePurge(queue, true)
	return err
}

func (m *Messenger) Close() error {
	return m.conn.Close()
}
