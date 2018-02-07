package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/encoding/yaml"
	"github.com/cuigh/auxo/errors"
)

type Source interface {
	Load() (data.Map, error)
	//Order() int
	//Update(notify func(data.Map))
}

//type flagSource struct {
//	flags *flag.FlagSet
//}
//
//func (s *flagSource) Load() (opts data.Map, err error) {
//	if s.flags == nil {
//		return
//	}
//
//	opts = make(data.Map)
//	s.flags.Visit(func(f *flag.Flag) {
//		getter := f.Value.(flag.Getter)
//		setOption(opts, f.Name, getter.Get())
//	})
//	return
//}
//
//func (s *flagSource) Default() (opts data.Map, err error) {
//	if s.flags == nil {
//		return
//	}
//
//	opts = make(data.Map)
//	set := make(map[string]struct{})
//	s.flags.Visit(func(f *flag.Flag) { set[f.Name] = struct{}{} })
//	s.flags.VisitAll(func(f *flag.Flag) {
//		if _, ok := set[f.Name]; !ok {
//			getter := f.Value.(flag.Getter)
//			setOption(opts, f.Name, getter.Get())
//		}
//	})
//	return
//}

type envSource struct {
	prefix  string
	aliases map[string]string
}

func (s *envSource) Load() (opts data.Map, err error) {
	opts = make(data.Map)
	envs := os.Environ()
	for _, env := range envs {
		opt := data.ParseOption(env, "=")
		key := opt.Name
		if s.prefix != "" {
			key = strings.TrimPrefix(key, s.prefix)
		}
		key = strings.Replace(strings.ToLower(opt.Name), "_", ".", -1)
		mergeOption(opts, key, opt.Value)
	}

	for alias, key := range s.aliases {
		if opt := os.Getenv(key); opt != "" {
			mergeOption(opts, alias, opt)
		}
	}
	return
}

func (s *envSource) SetPrefix(prefix string) {
	s.prefix = prefix
}

func (s *envSource) SetAlias(alias string, key string) {
	if s.aliases == nil {
		s.aliases = make(map[string]string)
	}
	s.aliases[alias] = key
}

type dataSource struct {
	Data []byte
	Type string
}

func (s *dataSource) Load() (data.Map, error) {
	return loadSource(s.Type, s.Data)
}

type fileSource string

func (fs fileSource) Load() (data.Map, error) {
	d, err := ioutil.ReadFile(string(fs))
	if err != nil {
		return nil, err
	}

	t := strings.ToLower(strings.TrimPrefix(filepath.Ext(string(fs)), "."))
	return loadSource(t, d)
}

func loadSource(t string, d []byte) (opts data.Map, err error) {
	opts = make(data.Map)
	switch t {
	case "yaml", "yml":
		err = yaml.Unmarshal(d, &opts)
	case "json":
		err = json.Unmarshal(d, &opts)
		//case "toml":
		//	err = toml.Unmarshal(d, &opts)
		//case "xml":
		//	err = xml.Unmarshal(d, &opts)
	default:
		return nil, errors.New("unsupported config type: " + t)
	}

	if err != nil {
		return nil, errors.Wrap(err, "loading config failed")
	}
	return opts, nil
}

func mergeOption(opts data.Map, k string, v interface{}) {
	keys := strings.Split(k, ".")
	last := len(keys) - 1
	for i, key := range keys {
		if opt, ok := opts[key]; ok {
			switch t := opt.(type) {
			case data.Map:
				opts = t
			case map[string]interface{}:
				opts = t
			default:
				return
			}
		} else {
			if i == last {
				opts[key] = v
			} else {
				m := data.Map{}
				opts[key] = m
				opts = m
			}
		}
	}
}

func coverOption(opts data.Map, k string, v interface{}) {
	keys := strings.Split(k, ".")
	last := len(keys) - 1
	for i, key := range keys {
		if opt, ok := opts[key]; ok {
			switch t := opt.(type) {
			case data.Map:
				opts = t
				continue
			case map[string]interface{}:
				opts = t
				continue
			}
		}

		if i == last {
			opts[key] = v
		} else {
			m := data.Map{}
			opts[key] = m
			opts = m
		}
	}
}
