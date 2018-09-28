package queueworker

import (
	"context"
	"errors"
	"testing"
	"time"

	internalTesting "github.com/Scalify/go-queue-worker/internal/testing"
	"github.com/streadway/amqp"
)

type test struct {
	Test bool
}

func newTestHandler(value interface{}, err error) MessageHandler {
	return func(msg amqp.Delivery) (interface{}, error) {
		return value, err
	}
}

func TestWorkerStartDeclaresQueues(t *testing.T) {
	q := internalTesting.NewTestQueue()
	b, logger := internalTesting.NewTestLogger()
	w, err := New(q, logger, 1, "job", "job-res", newTestHandler(nil, nil))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := w.Start(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()
	internalTesting.CheckLogger(t, b)

	if len(q.QueuesDeclared) != 2 {
		t.Fatalf("expected to have %d declared queues, got %d", 2, len(q.QueuesDeclared))
	}
}

func TestWorker(t *testing.T) {
	cases := []struct {
		value            interface{}
		queueDeclaredErr error
		consumeErr       error
		err, expErr      error
		queue            int
		acked, nacked    bool
	}{
		{
			value:            nil,
			err:              nil,
			expErr:           nil,
			queueDeclaredErr: nil,
			queue:            1,
			acked:            true,
			nacked:           false,
		},
		{
			value:            &test{false},
			err:              nil,
			expErr:           nil,
			queueDeclaredErr: nil,
			queue:            2,
			acked:            true,
			nacked:           false,
		},
		{
			value:            &test{false},
			err:              nil,
			expErr:           errors.New("unable to create queue job: whoops, something happened"),
			queueDeclaredErr: errors.New("whoops, something happened"),
			queue:            1,
			acked:            false,
			nacked:           false,
		},
		{
			value:      &test{false},
			err:        nil,
			expErr:     errors.New("failed to create queue consumer: whoops, something happened"),
			consumeErr: errors.New("whoops, something happened"),
			queue:      1,
			acked:      false,
			nacked:     false,
		},
	}

	for i, c := range cases {
		q := internalTesting.NewTestQueue()
		b, logger := internalTesting.NewTestLogger()
		w, newErr := New(q, logger, 1, "job", "job-res", newTestHandler(c.value, c.err))
		if newErr != nil {
			t.Fatalf("case %d: got error during new: %v", i, newErr)
		}

		q.Messages = append(q.Messages, []byte("{\"test\":true}"))
		q.QueueDeclaredErr = c.queueDeclaredErr
		q.ConsumeErr = c.consumeErr

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		startErr := w.Start(ctx)
		if c.expErr == nil && startErr != nil {
			t.Fatalf("case %d: got error during start: %v", i, startErr)
		} else if c.expErr != nil && (startErr == nil || c.expErr.Error() != startErr.Error()) {
			t.Fatalf("case %d: Expected to get error %q, got %q", i, c.expErr, startErr)
		}

		internalTesting.CheckLogger(t, b)

		if len(q.Messages) != c.queue {
			t.Fatalf("case %d: expected to have %d messages in queue, got %d", i, c.queue, len(q.Messages))
		}

		if c.nacked != q.Nacked {
			t.Fatalf("casde %d: Expected to be nacked = %v, but is %v", i, c.nacked, q.Nacked)
		}
		if c.acked != q.Acked {
			t.Fatalf("casde %d: Expected to be acked = %v, but is %v", i, c.acked, q.Acked)
		}
	}
}
