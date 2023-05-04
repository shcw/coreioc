package core

type NewInstance func(...any) (any, error)

type ServiceProvider interface {
	Name() string
	Register(Container) NewInstance
	Params(Container) []any
	IsDefer() bool
	Boot(Container) error
}
