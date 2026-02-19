package cacheindex

////////////////////////////////////////////////////////////////////////////////

type indexAlias struct {
	Index
	name string
}

func NewIndexAlias(name string, idx Index) Index {
	if name == idx.GetName() {
		return idx
	}
	return &indexAlias{idx, name}
}

func (i *indexAlias) GetEffective() Index {
	return i.Index.GetEffective()
}

func (i *indexAlias) GetName() string {
	return i.name
}
