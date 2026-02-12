package builder

type Builder interface {
	Named(name string) Builder
}
