package merge

import (
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/managedfields"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/openapi"
	"k8s.io/client-go/openapi3"
	"k8s.io/client-go/rest"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

type Converters interface {
	GetConverter(gvk schema.GroupVersionKind) (managedfields.TypeConverter, error)
}

type converters struct {
	lock       sync.Mutex
	client     openapi.Client
	converters map[schema.GroupVersion]managedfields.TypeConverter
}

func NewConverters(config *rest.Config) (Converters, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return &converters{
		client:     dc.OpenAPIV3(),
		converters: map[schema.GroupVersion]managedfields.TypeConverter{},
	}, nil
}

func (c *converters) GetConverter(gvk schema.GroupVersionKind) (managedfields.TypeConverter, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	req := gvk.GroupVersion()
	tc := c.converters[req]
	if tc != nil {
		return tc, nil
	}

	// 1. openapi3.NewRoot is the official helper to traverse the V3 discovery
	root := openapi3.NewRoot(c.client)

	definitions := make(map[string]*spec.Schema)

	// 2. Fetch the "Root" which contains all GroupVersions
	paths, err := c.client.Paths()
	if err != nil {
		return nil, err
	}

	for gvStr, _ := range paths {
		// Use the root helper to get the spec for each GroupVersion
		// This handles the network requests to /openapi/v3/apis/...
		if strings.HasPrefix(gvStr, "apis/") {
			gvStr = strings.TrimPrefix(gvStr, "apis/")
		} else if strings.HasPrefix(gvStr, "api/") {
			gvStr = strings.TrimPrefix(gvStr, "api/")
		}
		gv, err := schema.ParseGroupVersion(gvStr)
		if err != nil || gv != req {
			continue
		}
		spec3, err := root.GVSpec(gv)
		if err != nil {
			continue
		}

		// Convert OpenAPI V3 components to the V2 spec.Schema map for TypeConverter
		for name, s := range spec3.Components.Schemas {
			definitions[name] = s
		}
	}

	conv, err := managedfields.NewTypeConverter(definitions, false)
	if err != nil {
		return nil, err
	}
	c.converters[req] = conv
	return conv, nil
}
