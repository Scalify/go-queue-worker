package queueworker

import "github.com/streadway/amqp"

// Queue implements a basic AMQP client
type Queue interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Qos(prefetchCount, prefetchSize int, global bool) error
}

// MessageHandler
type MessageHandler func(msg amqp.Delivery) (interface{}, error)
