package storage

import (
	"github.com/mikaellindemann/registryfrontend"
	"github.com/mikaellindemann/registryfrontend/client"
	"strings"

	"github.com/pkg/errors"
)

type MemoryStorage map[string]registryfrontend.Registry

var (
	_                   registryfrontend.Storage = &MemoryStorage{}
	ErrIllegalName                               = errors.New("illegal character in registry name")
	ErrRegistryNotFound                          = errors.New("registry not found")
)

func NewInMemoryStorage() *MemoryStorage {
	s := make(MemoryStorage)
	return &s
}

func (m *MemoryStorage) Registries() ([]registryfrontend.Client, error) {
	res := make([]registryfrontend.Client, 0, len(*m))

	for _, reg := range *m {
		if reg.User != "" {
			c, err := client.MakeV2BasicAuth(reg.Name, reg.Url, reg.User, reg.Password)
			if err != nil {
				return nil, err
			}
			res = append(res, c)
		} else {
			c, err := client.MakeV2(reg.Name, reg.Url)
			if err != nil {
				return nil, err
			}
			res = append(res, c)
		}
	}

	return res, nil
}

func (m *MemoryStorage) Registry(name string) (registryfrontend.Client, error) {
	if reg, ok := (*m)[name]; !ok {
		return nil, ErrRegistryNotFound
	} else {
		if reg.User != "" {
			return client.MakeV2BasicAuth(reg.Name, reg.Url, reg.User, reg.Password)
		} else {
			return client.MakeV2(reg.Name, reg.Url)
		}
	}
}

func (m *MemoryStorage) Add(r registryfrontend.Registry) error {
	if hasIllegalCharacters(r.Name) {
		return ErrIllegalName
	}
	(*m)[r.Name] = r
	return nil
}

func (m *MemoryStorage) Update(r registryfrontend.Registry) error {
	return m.Add(r)
}

func (m *MemoryStorage) Clear() error {
	*m = make(map[string]registryfrontend.Registry)
	return nil
}

func (m *MemoryStorage) Remove(r registryfrontend.Registry) error {
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
