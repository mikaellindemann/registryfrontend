package registry_frontend

import (
	"strings"

	"github.com/pkg/errors"
)

type memoryStorage map[string]Registry

var (
	_                   Storage = &memoryStorage{}
	ErrIllegalName              = errors.New("illegal character in registry name")
	ErrRegistryNotFound         = errors.New("registry not found")
)

func NewInMemoryStorage() Storage {
	s := make(memoryStorage)
	return &s
}

func (m *memoryStorage) Registries() ([]Registry, error) {
	res := make([]Registry, 0, len(*m))

	for _, r := range *m {
		res = append(res, r)
	}

	return res, nil
}

func (m *memoryStorage) Registry(name string) (Registry, error) {
	if reg, ok := (*m)[name]; !ok {
		return reg, ErrRegistryNotFound
	} else {
		return reg, nil
	}
}

func (m *memoryStorage) Add(r Registry) error {
	if hasIllegalCharacters(r.Name) {
		return ErrIllegalName
	}
	(*m)[r.Name] = r
	return nil
}

func (m *memoryStorage) Update(r Registry) error {
	return m.Add(r)
}

func (m *memoryStorage) Clear() error {
	*m = make(map[string]Registry)
	return nil
}

func (m *memoryStorage) Remove(r Registry) error {
	delete(*m, r.Name)
	return nil
}

var chars = map[uint8]struct{}{
	'a': {},
	'b': {},
	'c': {},
	'd': {},
	'e': {},
	'f': {},
	'g': {},
	'h': {},
	'i': {},
	'j': {},
	'k': {},
	'l': {},
	'm': {},
	'n': {},
	'o': {},
	'p': {},
	'q': {},
	'r': {},
	's': {},
	't': {},
	'u': {},
	'v': {},
	'w': {},
	'x': {},
	'y': {},
	'z': {},
}

func hasIllegalCharacters(name string) bool {
	name = strings.ToLower(name)
	for i := range name {
		if _, ok := chars[name[i]]; !ok {
			return true
		}
	}
	return false
}
