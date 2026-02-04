package builder

import (
	"github.com/mandelsoft/kubecrtutils/controller/handler"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Builder interface {
	Named(name string) Builder
	Watches(object client.Object, eventHandler handler.EventHandler, opts ...WatchesOption) Builder
}
