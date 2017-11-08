package config

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Source interface {
	LoadData() ([]byte, error)
	GetType() string // yaml/toml/json...
}

type dataSource struct {
	Data []byte
	Type string
}

func (s *dataSource) LoadData() ([]byte, error) {
	return s.Data, nil
}

func (s *dataSource) GetType() string {
	return s.Type
}

type fileSource string

func (fs fileSource) GetType() string {
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(string(fs)), "."))
}

func (fs fileSource) LoadData() ([]byte, error) {
	return ioutil.ReadFile(string(fs))
}
