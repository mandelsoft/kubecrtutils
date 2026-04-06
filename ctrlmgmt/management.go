package ctrlmgmt

import (
	"context"
	"fmt"
	"slices"
	"sort"

	"github.com/mandelsoft/kubecrtutils/component"
	. "github.com/mandelsoft/kubecrtutils/log"
	"sigs.k8s.io/multicluster-runtime/pkg/manager"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/options/manageropts"
	"github.com/mandelsoft/logging"
	mcctrl "sigs.k8s.io/multicluster-runtime"
)

func NewControllerManagerByOpts(ctx context.Context, opts flagutils.OptionSetProvider) (ControllerManager, error) {
	def := From(opts)

	if def == nil {
		return nil, fmt.Errorf("no management definition found")
	}

	if def.GetError() != nil {
		return nil, fmt.Errorf("management definition: %w", def.GetError())
	}

	mopts := manageropts.From(opts)
	if mopts == nil {
		return nil, fmt.Errorf("no manager options found")
	}

	copts := cluster.From(opts)
	if copts == nil {
		return nil, fmt.Errorf("no clusters found in options")
	}
	clusters := copts.GetClusters()

	logger := kubecrtutils.LogContext.WithContext(logging.NewRealm(kubecrtutils.Realm.Name() + "/" + def.GetName())).Logger()
	logger.Info("configure controller manager {{cm}}", "cm", def.GetName())

	list := []string{}
	defcluster := false
	for n := range clusters.Elements {
		if n == cluster.DEFAULT {
			defcluster = true
		} else {
			list = append(list, n)
		}
	}
	sort.Strings(list)
	if defcluster {
		list = slices.Insert(list, 0, cluster.DEFAULT)
	}

	main := clusters.Get(mopts.GetMain())
	if main.AsFleet() != nil {
		main = main.AsFleet().GetBaseCluster()
		if main == nil {
			return nil, fmt.Errorf("no fleet base cluster usable as main cluster for controller manager")
		}
	}

	for _, n := range list {
		c := clusters.Get(n)
		if c == c.GetEffective() {
			Info(logger, "  using configured ", ClusterInfo(c))
		} else {
			Info(logger, "  using logical ", LogicalClusterInfo(c))
		}
	}

	mcmgr, err := mopts.GetManager(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("settingup manager: %w", err)
	}

	cntropts := controller.From(opts)
	if cntropts == nil {
		return nil, fmt.Errorf("no controller definitions found")
	}

	cm := &_controllermanager{
		Element:    internal.NewElement(def.GetName()),
		logger:     logger,
		clusters:   clusters,
		manager:    mcmgr,
		main:       main.AsCluster(),
		indices:    cacheindex.NewIndices(),
		components: component.NewComponents(),
		definition: def,
	}

	iopts := cacheindex.From(opts)
	if iopts != nil && iopts.Len() > 0 {
		logger.Info("configure global indices...")
		err = iopts.CreateIndices(ctx, nil, cm)
		if err != nil {
			return nil, fmt.Errorf("settingup indices: %w", err)
		}
	}

	coopts := component.From(opts)
	if coopts != nil && coopts.Len() > 0 {
		logger.Info("configure component indices...")
		err = coopts.CreateIndices(ctx, nil, cm)
		if err != nil {
			return nil, fmt.Errorf("setting up component indices: %w", err)
		}
	}

	if cntropts != nil && cntropts.Len() > 0 {
		logger.Info("configure controller indices...")
		err = cntropts.CreateIndices(ctx, nil, cm)
		if err != nil {
			return nil, fmt.Errorf("setting up controller indices: %w", err)
		}
	}

	if coopts != nil && coopts.Len() > 0 {
		err = coopts.Apply(ctx, nil, cm)
		if err != nil {
			return nil, fmt.Errorf("setting up components: %w", err)
		}

		for _, c := range cm.components.Elements {
			if r, ok := c.(manager.Runnable); ok {
				logger.Info("registering component {{comp}} at manager", "comp", c.GetName())
				err := mcmgr.Add(r)
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		cm.components = component.NewComponents()
	}

	cm.controllers = controller.NewControllers()
	err = cntropts.Apply(ctx, nil, cm)
	if err != nil {
		return nil, fmt.Errorf("settingup controllers: %w", err)
	}
	return cm, nil
}

type _controllermanager struct {
	internal.Element
	logger      logging.Logger
	main        cluster.Cluster
	manager     mcctrl.Manager
	clusters    cluster.Clusters
	indices     cacheindex.Indices
	controllers controller.Controllers
	components  component.Components
	definition  Definition
}

func (cm *_controllermanager) GetLogger() logging.Logger {
	return cm.logger
}

func (cm *_controllermanager) GetControllerDefinition(name string) controller.Definition {
	return cm.definition.GetController(name)
}

func (cm *_controllermanager) GetComponents() component.Components {
	return cm.components
}

func (cm *_controllermanager) GetControllers() controller.Controllers {
	return cm.controllers
}

func (cm *_controllermanager) GetClusters() cluster.Clusters {
	return cm.clusters
}

func (cm *_controllermanager) GetIndices() cacheindex.Indices {
	return cm.indices
}

func (cm *_controllermanager) GetManager() mcctrl.Manager {
	return cm.manager
}

func (cm *_controllermanager) GetMainCluster() cluster.ClusterEquivalent {
	return cm.main
}

func (cm *_controllermanager) MapTechnicalName(name string) cluster.ClusterEquivalent {
	return cm.clusters.Get(name)
}

func (cm *_controllermanager) GetIndex(name string) cluster.Index {
	return cm.indices.Get(name)
}

func (cm *_controllermanager) GetComponent(name string) component.Component {
	return cm.components.Get(name)
}
