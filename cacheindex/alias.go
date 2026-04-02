package cacheindex

type definitionAlias struct {
	Definition
	name string
}

func NewDefinitionAlias(name string, idx Definition) Definition {
	if name == idx.GetName() {
		return idx
	}
	return &definitionAlias{idx, name}
}

func (i *definitionAlias) GetEffective() Definition {
	return i.Definition.GetEffective()
}

func (i *definitionAlias) GetName() string {
	return i.name
}

////////////////////////////////////////////////////////////////////////////////

type indexAlias struct {
	Index
	name string
}

func NewAlias(name string, idx Index) Index {
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
