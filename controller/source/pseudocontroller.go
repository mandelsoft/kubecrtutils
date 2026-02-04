/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller/priorityqueue"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// watches implements controller.Watches and corresponds to a basic
// controller without explicit reconcilation element.
type watches[request comparable] struct {
	// Name is used to uniquely identify a watches in tracing, logging and monitoring.  Name is required.
	Name string

	Logger logr.Logger

	// mu is used to synchronize watches setup
	mu sync.Mutex

	// Started is true if the watches has been Started
	Started bool

	// ctx is the context that was passed to Start() and used when starting watches.
	//
	// According to the docs, contexts should not be stored in a struct: https://golang.org/pkg/context,
	// while we usually always strive to follow best practices, we consider this a legacy case and it should
	// undergo a major refactoring and redesign to allow for context to not be stored in a struct.
	ctx context.Context

	// CacheSyncTimeout refers to the time limit set on waiting for cache to sync
	// Defaults to 2 minutes if not set.
	CacheSyncTimeout time.Duration

	// startWatches maintains a list of sources, handlers, and predicates to start when the controller is started.
	startWatches []source.TypedSource[request]

	// startedEventSourcesAndQueue is used to track if the event sources have been started.
	// It ensures that we append sources to c.startWatches only until we call Start() / Warmup()
	// It is true if startEventSourcesAndQueueLocked has been called at least once.
	startedEventSourcesAndQueue bool

	// didStartEventSourcesOnce is used to ensure that the event sources are only started once.
	didStartEventSourcesOnce sync.Once

	// EnableWarmup specifies whether the controller should start its sources when the manager is not
	// the leader. This is useful for cases where sources take a long time to start, as it allows
	// for the controller to warm up its caches even before it is elected as the leader. This
	// improves leadership failover time, as the caches will be prepopulated before the controller
	// transitions to be leader.
	//
	// Setting EnableWarmup to true and NeedLeaderElection to true means the controller will start its
	// sources without waiting to become leader.
	// Setting EnableWarmup to true and NeedLeaderElection to false is a no-op as controllers without
	// leader election do not wait on leader election to start their sources.
	// Defaults to false.
	EnableWarmup *bool
}

// Watch implements controller.watches.
func (c *watches[request]) Watch(src source.TypedSource[request]) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Sources weren't started yet, store the watches locally and return.
	// These sources are going to be held until either Warmup() or Start(...) is called.
	if !c.startedEventSourcesAndQueue {
		c.startWatches = append(c.startWatches, src)
		return nil
	}

	return src.Start(c.ctx, nil)
}

// Warmup implements the manager.WarmupRunnable interface.
func (c *watches[request]) Warmup(ctx context.Context) error {
	if c.EnableWarmup == nil || !*c.EnableWarmup {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Set the ctx so later calls to watch use this internal context
	c.ctx = ctx

	return c.startEventSourcesAndQueueLocked(ctx)
}

// Start implements controller.watches.
func (c *watches[request]) Start(ctx context.Context) error {
	// use an IIFE to get proper lock handling
	// but lock outside to get proper handling of the queue shutdown
	c.mu.Lock()
	if c.Started {
		return errors.New("watcher was started more than once. This is likely to be caused by being added to a manager multiple times")
	}

	// Set the internal context.
	c.ctx = ctx

	wg := &sync.WaitGroup{}
	err := func() error {
		defer c.mu.Unlock()

		// TODO(pwittrock): Reconsider HandleCrash
		defer utilruntime.HandleCrashWithLogger(c.Logger)

		// NB(directxman12): launch the sources *before* trying to wait for the
		// caches to sync so that they have a chance to register their intended
		// caches.
		if err := c.startEventSourcesAndQueueLocked(ctx); err != nil {
			return err
		}

		c.Started = true
		return nil
	}()
	if err != nil {
		return err
	}

	<-ctx.Done()
	wg.Wait()
	return nil
}

// startEventSourcesAndQueueLocked launches all the sources registered with this controller and waits
// for them to sync. It returns an error if any of the sources fail to start or sync.
func (c *watches[request]) startEventSourcesAndQueueLocked(ctx context.Context) error {
	var retErr error

	c.didStartEventSourcesOnce.Do(func() {
		go func() {
			<-ctx.Done()
		}()

		errGroup := &errgroup.Group{}
		for _, watch := range c.startWatches {
			didStartSyncingSource := &atomic.Bool{}
			errGroup.Go(func() error {
				// Use a timeout for starting and syncing the source to avoid silently
				// blocking startup indefinitely if it doesn't come up.
				sourceStartCtx, cancel := context.WithTimeout(ctx, c.CacheSyncTimeout)
				defer cancel()

				sourceStartErrChan := make(chan error, 1) // Buffer chan to not leak goroutine if we time out
				go func() {
					defer close(sourceStartErrChan)

					if err := watch.Start(ctx, nil); err != nil {
						sourceStartErrChan <- err
						return
					}
					syncingSource, ok := watch.(source.TypedSyncingSource[request])
					if !ok {
						return
					}
					didStartSyncingSource.Store(true)
					if err := syncingSource.WaitForSync(sourceStartCtx); err != nil {
						err := fmt.Errorf("failed to wait for %s caches to sync %v: %w", c.Name, syncingSource, err)
						sourceStartErrChan <- err
					}
				}()

				select {
				case err := <-sourceStartErrChan:
					return err
				case <-sourceStartCtx.Done():
					if didStartSyncingSource.Load() { // We are racing with WaitForSync, wait for it to let it tell us what happened
						return <-sourceStartErrChan
					}
					if ctx.Err() != nil { // Don't return an error if the root context got cancelled
						return nil
					}
					return fmt.Errorf("timed out waiting for source %s to Start. Please ensure that its Start() method is non-blocking", watch)
				}
			})
		}
		retErr = errGroup.Wait()

		// All the watches have been started, we can reset the local slice.
		//
		// We should never hold watches more than necessary, each watch source can hold a backing cache,
		// which won't be garbage collected if we hold a reference to it.
		c.startWatches = nil

		// Mark event sources as started after resetting the startWatches slice so that watches from
		// a new Watch() call are immediately started.
		c.startedEventSourcesAndQueue = true
	})

	return retErr
}

type priorityQueueWrapper[request comparable] struct {
	workqueue.TypedRateLimitingInterface[request]
}

func (p *priorityQueueWrapper[request]) AddWithOpts(opts priorityqueue.AddOpts, items ...request) {
	for _, item := range items {
		switch {
		case opts.RateLimited:
			p.TypedRateLimitingInterface.AddRateLimited(item)
		case opts.After > 0:
			p.TypedRateLimitingInterface.AddAfter(item, opts.After)
		default:
			p.TypedRateLimitingInterface.Add(item)
		}
	}
}

func (p *priorityQueueWrapper[request]) GetWithPriority() (request, int, bool) {
	item, shutdown := p.TypedRateLimitingInterface.Get()
	return item, 0, shutdown
}
