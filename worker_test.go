package queueworker

import (
	"context"
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
		value  interface{}
		err    error
		queue  int
		expErr error
		acked, nacked bool
	}{
		{
			value:  nil,
			err:    nil,
			queue:  1,
			acked: true,
			nacked: false,
		},
		{
			value:  &test{false},
			err:    nil,
			queue:  2,
			acked: true,
			nacked: false,
		},
	}

	for i, c := range cases {
		q := internalTesting.NewTestQueue()
		b, logger := internalTesting.NewTestLogger()
		w, err := New(q, logger, 1, "job", "job-res", newTestHandler(c.value, c.err))
		if err != nil {
			t.Fatal(err)
		}

		q.Messages = append(q.Messages, []byte("{\"test\":true}"))

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			if err := w.Start(ctx); err != nil {
				t.Fatal(err)
			}
		}()

		time.Sleep(10 * time.Millisecond)
		cancel()
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
