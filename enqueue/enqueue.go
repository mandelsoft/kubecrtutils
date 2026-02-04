package enqueue

import (
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Enqueue = TypedEnqueue[reconcile.Request]

var _ Enqueue = (*typedenqueue[reconcile.Request])(nil)

func NewEnqueue() Enqueue {
	return NewTypedEnqueue[reconcile.Request]()
}
