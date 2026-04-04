package component

// GetName also has to map its own name based on the
// defined component mappings.
func (d *_mapped) GetName() string {
	return d.ComponentMappings().Map(d.Definition.GetName())
}
