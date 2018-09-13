package queueworker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

func (w *Worker) ensureQueues() error {
	var err error
	var queues = []string{w.queueNameJobs, w.queueNameJobResults}

	for _, queueName := range queues {
		_, err = w.queue.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("unable to create queue %s: %v", queueName, err)
		}

		w.logger.Debugf("Checked queue %s for existence.", queueName)
	}

	return nil
}

func (w *Worker) publishResult(result interface{}) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return w.queue.Publish("", w.queueNameJobResults, false, false, amqp.Publishing{
		ContentType: contentTypeJSON,
		Body:        b,
	})
}

func (w *Worker) consumeJobs(ctx context.Context) error {
	if err := w.queue.Qos(int(w.parallelExecutions), 0, false); err != nil {
		w.logger.Fatalf("Failed to set queue QOS: %v", err)
	}

	consumer, err := w.queue.Consume(w.queueNameJobs, w.consumerTag, false, false, false, false, nil)
	if err != nil {
		w.logger.Fatalf("Failed to create queue consumer: %v", err)
		return err
	}

	var msg amqp.Delivery

	for {
		select {
		case msg = <-consumer:
			if string(msg.Body) == "" {
				continue
			}

			go w.handleMessage(msg)
		case <-ctx.Done():
			return nil
		}
	}
}
func (w *Worker) handleMessage(msg amqp.Delivery) {
	res, err := w.messageHandler(msg)
	if err != nil {
		if errNack := msg.Nack(false, true); errNack != nil {
			w.logger.Errorf("Failed to nack message: %v", errNack)
		}
		return
	}

	if res != nil {
		if err := w.publishResult(res); err != nil {
			w.logger.Errorf("Failed to publish job result: %v", err)
			return
		}
	}

	if errAck := msg.Ack(false); errAck != nil {
		w.logger.Errorf("Failed to ack message: %v", errAck)
	}
}
