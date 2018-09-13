package queueworker

import (
	"context"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
)

// Worker executes jobs which are consumed from a queue
type Worker struct {
	queue               Queue
	logger              *logrus.Entry
	parallelExecutions  uint
	queueNameJobs       string
	queueNameJobResults string
	consumerTag         string
	messageHandler      MessageHandler
}

// New returns a new Worker instance
func New(queue Queue, logger *logrus.Entry, parallelExecutions uint, queueNameJobs, queueNameJobResults string, handler MessageHandler) (*Worker, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate consumer tag UUID: %v", err)
	}
	return &Worker{
		queue:               queue,
		logger:              logger,
		parallelExecutions:  parallelExecutions,
		queueNameJobs:       queueNameJobs,
		queueNameJobResults: queueNameJobResults,
		consumerTag:         id.String(),
		messageHandler:      handler,
	}, nil
}

// Start the consumer
func (w *Worker) Start(ctx context.Context) error {
	if err := w.ensureQueues(); err != nil {
		return err
	}

	return w.consumeJobs(ctx)
}
