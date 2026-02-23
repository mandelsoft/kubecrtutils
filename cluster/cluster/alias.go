package cluster

type _clusterAlias struct {
	Cluster
	name string
}

func NewAlias(name string, c Cluster) Cluster {
	if name == c.GetName() {
		return c
	}
	return &_clusterAlias{
		Cluster: c,
		name:    name,
	}
}

func (c *_clusterAlias) GetName() string {
	return c.name
}

func (c *_clusterAlias) Unwrap() Cluster {
	return c.Cluster
}

func (c *_clusterAlias) LiftTechnical(clusterName string) (string, Cluster) {
	c.Cluster.LiftTechnical(clusterName) // panic to indicate  corruption
	return c.name, c
}
