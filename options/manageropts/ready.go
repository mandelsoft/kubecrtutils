package manageropts

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mandelsoft/goutils/atomic"
	"github.com/mandelsoft/kubecrtutils/ctxutils"
	ctrl "sigs.k8s.io/multicluster-runtime"
)

type CacheWait interface {
	WaitForCacheSync(ctx context.Context) bool
}
type Ready struct {
	err atomic.InterfaceValue[error]

	mgr     ctrl.Manager
	waiters map[string]CacheWait
}

func NewReady(mgr ctrl.Manager) *Ready {
	r := &Ready{mgr: mgr, waiters: make(map[string]CacheWait)}
	r.mgr.GetLocalManager().Add(r)
	return r
}

func (ready *Ready) Add(name string, wait CacheWait) {
	ready.waiters[name] = wait
}

func (ready *Ready) GetState() error {
	return ready.err.Load()
}

func (ready *Ready) Check(r *http.Request) error {
	return ready.GetState()
}

func (ready *Ready) Start(ctx context.Context) error {
	go func() {
		var err error
		ready.mgr.GetLogger().Info("starting cache synced checker")
		for n, w := range ready.waiters {
			ready.mgr.GetLogger().Info("waiting for {{target}}", "target", n)
			ready.err.Store(fmt.Errorf("waiting for %s", n))
			if !w.WaitForCacheSync(ctx) {
				err = fmt.Errorf("failed to get cache for %s ready", n)
				ready.mgr.GetLogger().Error(err, "abort")
				ready.err.Store(err)
				ctxutils.Cancel(ready.mgr, ctx)
			}
			ready.mgr.GetLogger().Info("{{target}} cache synched", "target", n)
		}
		err = nil
		ready.err.Store(err)
	}()
	return nil
}
