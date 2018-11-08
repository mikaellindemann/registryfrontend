package registryfrontend

import (
	"strings"

	"github.com/pkg/errors"
)

type MemoryStorage map[string]Registry

var (
	_                   Storage = &MemoryStorage{}
	ErrIllegalName              = errors.New("illegal character in registry name")
	ErrRegistryNotFound         = errors.New("registry not found")
)

func NewInMemoryStorage() *MemoryStorage {
	s := make(MemoryStorage)
	return &s
}

func (m *MemoryStorage) Registries() ([]Registry, error) {
	res := make([]Registry, 0, len(*m))

	for _, r := range *m {
		res = append(res, r)
	}

	return res, nil
}

func (m *MemoryStorage) Registry(name string) (Registry, error) {
	if reg, ok := (*m)[name]; !ok {
		return reg, ErrRegistryNotFound
	} else {
		return reg, nil
	}
}

func (m *MemoryStorage) Add(r Registry) error {
	if hasIllegalCharacters(r.Name) {
		return ErrIllegalName
	}
	(*m)[r.Name] = r
	return nil
}

func (m *MemoryStorage) Update(r Registry) error {
	return m.Add(r)
}

func (m *MemoryStorage) Clear() error {
	*m = make(map[string]Registry)
	return nil
}

func (m *MemoryStorage) Remove(r Registry) error {
	delete(*m, r.Name)
	return nil
}

func hasIllegalCharacters(name string) bool {
	name = strings.ToLower(name)
	for i := range name {
		if !((name[i] >= 'a' && name[i] <= 'z') || (name[i] >= '0' && name[i] <= '9')) {
			return true
		}
	}
	return false
}
