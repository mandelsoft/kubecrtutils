package cacheindex

type _alias struct {
	Index
	name string
}

func NewAlias(name string, idx Index) Index {
	if name == idx.GetName() {
		return idx
	}
	return &_alias{idx, name}
}

func (i *_alias) GetEffective() Index {
	return i.Index.GetEffective()
}

func (i *_alias) GetName() string {
	return i.name
}
