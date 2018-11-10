package registryfrontend

type Storage interface {
	Registries() ([]Client, error)
	Registry(name string) (Client, error)
	Add(r Registry) error
	Update(r Registry) error
	Clear() error
	Remove(r Registry) error
}
