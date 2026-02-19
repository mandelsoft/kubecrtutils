package ctrlmgmt

import (
	"context"
	"fmt"
	"slices"
	"sort"

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
		typ := "cluster"
		if c.AsFleet() != nil {
			typ = "fleet"
		}
		if c == c.GetEffective() {
			logger.Info("using configured {{type}} {{cluster}}[{{identity}}] accessing {{info}}", "type", typ, "cluster", n, "effective", c.GetEffective().GetName(), "identity", c.GetId(), "info", c.GetInfo())
		} else {
			logger.Info("using logical {{type}} {{cluster}} mapped to {{effective}}", "type", typ, "cluster", n, "effective", c.GetEffective().GetName())
		}
	}

	manager, err := mopts.GetManager(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("settingup manager: %w", err)
	}

	cntropts := controller.From(opts)
	if cntropts == nil {
		return nil, fmt.Errorf("no controller definitions found")
	}

	var indices cacheindex.Indices
	iopts := cacheindex.From(opts)
	if iopts != nil && iopts.Len() > 0 {
		logger.Info("configure global indices...")
		indices, err = iopts.GetIndices(ctx, clusters, logger)
		if err != nil {
			return nil, fmt.Errorf("settingup indices: %w", err)
		}
	} else {
		indices = cacheindex.NewIndices()
	}

	cm := &_controllermanager{
		Element:    internal.NewElement(def.GetName()),
		logger:     logger,
		clusters:   clusters,
		manager:    manager,
		main:       main.AsCluster(),
		indices:    indices,
		definition: def,
	}

	cntr, err := cntropts.Apply(ctx, cm)
	if err != nil {
		return nil, fmt.Errorf("settingup controllers: %w", err)
	}
	cm.controllers = cntr
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
	definition  Definition
}

func (cm *_controllermanager) GetLogger() logging.Logger {
	return cm.logger
}

func (cm *_controllermanager) GetControllerDefinition(name string) controller.Definition {
	return cm.definition.GetController(name)
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
