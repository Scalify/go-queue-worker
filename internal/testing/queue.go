package testing

import "github.com/streadway/amqp"

// TestQueue for tests which buffers the messages
type TestQueue struct {
	NoopAcknowledger

	QueuesDeclared []string
	Messages       [][]byte
}

// NewTestQueue returns a new TestQueue instance
func NewTestQueue() *TestQueue {
	return &TestQueue{
		QueuesDeclared: make([]string, 0),
		Messages:       make([][]byte, 0),
	}
}

// QueueDeclare doesn't really do something.
func (t *TestQueue) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	t.QueuesDeclared = append(t.QueuesDeclared, name)
	return amqp.Queue{
		Name: name,
	}, nil
}

// Consume just pushes all messages from the Messages field into the channel.
func (t *TestQueue) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	c := make(chan amqp.Delivery, len(t.Messages))

	for _, msg := range t.Messages {
		c <- amqp.Delivery{
			Body:         msg,
			Acknowledger: t,
		}
	}

	return c, nil
}

// Publish adds the given message to the Messages field.
func (t *TestQueue) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	t.Messages = append(t.Messages, msg.Body)

	return nil
}

// Qos doesn't really do something.
func (t *TestQueue) Qos(prefetchCount, prefetchSize int, global bool) error {
	return nil
}

// NoopAcknowledger doesn't really do something.
type NoopAcknowledger struct {
	Nacked, Acked bool
}

// Ack doesn't do something.
func (a *NoopAcknowledger) Ack(tag uint64, multiple bool) error {
	a.Acked = true
	return nil
}

// Nack doesn't do something.
func (a *NoopAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	a.Nacked = true
	return nil
}

// Reject doesn't do something.
func (a *NoopAcknowledger) Reject(tag uint64, requeue bool) error {
	a.Nacked = true
	return nil
}
