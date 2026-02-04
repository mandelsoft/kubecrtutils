package controller

import (
	"context"
	"fmt"
	"sync"

	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/client-go/util/workqueue"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type TypedSource[object client.Object, request comparable] interface {
	ForCluster(string, cluster.Cluster) (source.TypedSource[request], bool, error)
}

////////////////////////////////////////////////////////////////////////////////

type ShadowController[request mcreconcile.ClusterAware[request]] interface {
	MultiClusterWatch(src TypedSource[client.Object, request]) error
}

func NewShadowController[request mcreconcile.ClusterAware[request]](cntr types.Controller) (ShadowController[request], error) {
	c := &mcSources[request]{
		clusters: make(map[string]*engagedCluster),
	}

	c.watches.Logger = cntr.GetLogger().WithName("watches").V(0)
	c.Logger.Info("creating shaddow controller for watch sources")
	// Add as a Manager components
	return c, cntr.GetControllerManager().GetManager().Add(c)
}

type mcSources[request comparable] struct {
	watches[request]
	lock     sync.Mutex
	clusters map[string]*engagedCluster
	sources  []TypedSource[client.Object, request]
}

type engagedCluster struct {
	name    string
	cluster cluster.Cluster
	ctx     context.Context
	cancel  context.CancelFunc
}

func (c *mcSources[request]) Engage(ctx context.Context, name string, cl cluster.Cluster) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Check if we already have this cluster engaged with the SAME context
	if old, ok := c.clusters[name]; ok {
		if old.cluster == cl && old.ctx.Err() == nil {
			// Same impl, engagement still live → nothing to do
			return nil
		}
		// Re-engage: either old ctx is done, or impl changed. Stop the old one if still live.
		if old.ctx.Err() == nil {
			old.cancel()
		}
		delete(c.clusters, name)
	}

	engCtx, cancel := context.WithCancel(ctx)

	// engage cluster aware instances
	for _, aware := range c.sources {
		src, shouldEngage, err := aware.ForCluster(name, cl)
		if err != nil {
			cancel()
			return fmt.Errorf("failed to engage for cluster %q: %w", name, err)
		}
		if !shouldEngage {
			continue
		}
		if err := c.watches.Watch(startWithinContext(engCtx, src)); err != nil {
			cancel()
			return fmt.Errorf("failed to watch for cluster %q: %w", name, err)
		}
	}

	ec := &engagedCluster{
		name:    name,
		cluster: cl,
		ctx:     engCtx,
		cancel:  cancel,
	}
	c.clusters[name] = ec
	go func(ctx context.Context, key string, token *engagedCluster) {
		<-ctx.Done()
		c.lock.Lock()
		defer c.lock.Unlock()
		if cur, ok := c.clusters[key]; ok && cur == token {
			delete(c.clusters, key)
		}
		// note: cancel() is driven by parent; no need to call here
	}(engCtx, name, ec)

	return nil
}

func (c *mcSources[request]) MultiClusterWatch(src TypedSource[client.Object, request]) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	for name, eng := range c.clusters {
		src, shouldEngage, err := src.ForCluster(name, eng.cluster)
		if err != nil {
			return fmt.Errorf("failed to engage for cluster %q: %w", name, err)
		}
		if !shouldEngage {
			continue
		}
		if err := c.watches.Watch(startWithinContext[request](eng.ctx, src)); err != nil {
			return fmt.Errorf("failed to watch for cluster %q: %w", name, err)
		}
	}

	c.sources = append(c.sources, src)

	return nil
}

func startWithinContext[request comparable](ctx context.Context, src source.TypedSource[request]) source.TypedSource[request] {
	return source.TypedFunc[request](func(ctlCtx context.Context, w workqueue.TypedRateLimitingInterface[request]) error {
		ctx, cancel := context.WithCancel(ctx)
		go func() {
			<-ctlCtx.Done()
			cancel()
		}()
		return src.Start(ctx, w)
	})
}
