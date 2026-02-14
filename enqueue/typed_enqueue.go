package enqueue

import (
	"context"
	"sync"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type TypedEnqueue[T comparable] interface {
	source.TypedSource[T]
	AddToQueue(req T)
}

type typedenqueue[T comparable] struct {
	lock   sync.Mutex
	queues []workqueue.TypedRateLimitingInterface[T]
}

var _ TypedEnqueue[int] = (*typedenqueue[int])(nil)

func NewTypedEnqueue[T comparable]() TypedEnqueue[T] {
	return &typedenqueue[T]{}
}

func (e *typedenqueue[T]) Start(ctx context.Context, w workqueue.TypedRateLimitingInterface[T]) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	for _, q := range e.queues {
		if w == q {
			return nil
		}
	}
	e.queues = append(e.queues, w)
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (e *typedenqueue[T]) AddToQueue(req T) {
	e.lock.Lock()
	defer e.lock.Unlock()

	for _, q := range e.queues {
		q.AddRateLimited(req)
	}
}
