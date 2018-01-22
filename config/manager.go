// Package config manages configurations of application.
package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/encoding/yaml"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/files"
	"github.com/cuigh/auxo/util/cast"
)

var (
	exts            = []string{".yml", ".yaml", ".toml", ".json"}
	unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
)

// Unmarshaler is custom unmarshal interface for Config.Unmarshal method.
type Unmarshaler interface {
	Unmarshal(i interface{}) error
}

type Manager struct {
	locker   sync.Mutex
	loaded   bool
	autoLoad bool

	// flags
	flags *flag.FlagSet

	// env
	envPrefix string
	env       map[string]string

	// options(local/source)
	options  data.Map
	profiles []string
	dirs     []string
	name     string
	srcs     []Source
	//mgrs     []SourceManager

	// defaults
	defaults data.Map
}

func New(name ...string) *Manager {
	m := &Manager{}
	if len(name) > 0 {
		m.SetName(name[0])
	} else {
		m.SetName("app")
	}
	return m
}

// FindFile searches all config directories and return the first found file.
func (m *Manager) FindFile(name string, exts ...string) string {
	for _, dir := range m.dirs {
		for _, ext := range exts {
			p := filepath.Join(dir, name+ext)
			if files.Exist(p) {
				return p
			}
		}
	}
	return ""
}

// FindFile searches all config directories and return all found files.
func (m *Manager) FindFiles(name string, exts ...string) []string {
	var list []string
	for _, dir := range m.dirs {
		for _, ext := range exts {
			p := filepath.Join(dir, name+ext)
			if files.Exist(p) {
				list = append(list, p)
			}
		}
	}
	return list
}

// FindFolder searches all config directories and return the first found folder.
func (m *Manager) FindFolder(name string) string {
	for _, dir := range m.dirs {
		p := filepath.Join(dir, name)
		if files.Exist(p) {
			return p
		}
	}
	return ""
}

// FindFolders searches all config directories and return all found folders.
func (m *Manager) FindFolders(name string) []string {
	var list []string
	for _, dir := range m.dirs {
		p := filepath.Join(dir, name)
		if files.Exist(p) {
			list = append(list, p)
		}
	}
	return list
}

// BindFlags binds a flag set.
func (m *Manager) BindFlags(set *flag.FlagSet) {
	m.flags = set
}

// SetEnvPrefix sets the prefix of environment variables. Default prefix is "AUXO".
func (m *Manager) SetEnvPrefix(prefix string) {
	m.envPrefix = prefix
}

// BindEnv binds a option to an environment variable.
func (m *Manager) BindEnv(key string, envKey string) {
	if m.env == nil {
		m.env = make(map[string]string)
	}
	m.env[key] = envKey
}

// SetProfiles sets active profiles. Profiles are only valid to local file sources.
func (m *Manager) SetProfile(profiles ...string) {
	m.profiles = profiles
}

// SetName sets name of the main configuration file (without extension).
func (m *Manager) SetName(name string) {
	m.name = name
}

// AddFolders adds directories to searching list.
func (m *Manager) AddFolder(dirs ...string) {
	m.dirs = append(m.dirs, dirs...)
}

// AddSource adds custom configuration sources.
func (m *Manager) AddSource(srcs ...Source) {
	m.srcs = append(m.srcs, srcs...)
}

// AddDataSource add a config source with bytes and type.
func (m *Manager) AddDataSource(data []byte, ct string) {
	src := &dataSource{
		Data: data,
		Type: ct,
	}
	m.AddSource(src)
}

func (m *Manager) addDefaultFolders() {
	exec, err := os.Executable()
	if err != nil {
		panic(err)
	}

	dir := filepath.Join(filepath.Dir(exec), "config")
	if files.NotExist(dir) {
		wd, err := os.Getwd()
		if err == nil {
			dir = filepath.Join(wd, "config")
		}
	}
	if files.Exist(dir) {
		m.AddFolder(dir)
	}
}

// SetDefaultValue sets a default option.
func (m *Manager) SetDefaultValue(name string, value interface{}) {
	if m.defaults == nil {
		m.defaults = data.Map{}
	}
	m.defaults.Set(name, value)
}

// Load reads options from all sources.
func (m *Manager) Load() error {
	return m.load(true)
}

func (m *Manager) load(force bool) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	if m.loaded && !force {
		return nil
	}

	m.options = data.Map{}

	m.loadFlags()
	m.loadEnvs()

	// file sources
	srcs := m.findFileSources()
	err := m.loadSources(srcs)
	if err != nil {
		return err
	}

	// custom source
	err = m.loadSources(m.srcs)
	if err != nil {
		return err
	}

	m.loaded = true
	return nil
}

func (m *Manager) loadFlags() {
	if m.flags != nil {
		m.flags.VisitAll(func(f *flag.Flag) {
			getter := f.Value.(flag.Getter)
			m.set(f.Name, getter.Get())
		})
	}
}

func (m *Manager) loadEnvs() {
	envs := os.Environ()
	for _, env := range envs {
		opt := data.ParseOption(env, "=")
		key := opt.Name
		if m.envPrefix != "" {
			key = strings.TrimPrefix(key, m.envPrefix)
		}
		key = strings.Replace(strings.ToLower(opt.Name), "_", ".", -1)
		m.set(key, opt.Value)
	}

	for key, envKey := range m.env {
		opt := os.Getenv(envKey)
		m.set(key, opt)
	}
}

func (m *Manager) set(k string, v interface{}) {
	opts := m.options
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
				break
			}
		} else {
			if i == last {
				opts[key] = v
			} else {
				t := data.Map{}
				opts[key] = t
				opts = t
			}
		}
	}
}

func (m *Manager) findFileSources() (srcs []Source) {
	if len(m.dirs) == 0 {
		m.addDefaultFolders()
	}

	for _, dir := range m.dirs {
		for _, ext := range exts {
			for _, profile := range m.profiles {
				path := filepath.Join(dir, m.name+"."+profile+ext)
				if files.Exist(path) {
					srcs = append(srcs, fileSource(path))
				}
			}
		}
	}
	for _, dir := range m.dirs {
		for _, ext := range exts {
			path := filepath.Join(dir, m.name+ext)
			if files.Exist(path) {
				srcs = append(srcs, fileSource(path))
			}
		}
	}
	return
}

func (m *Manager) loadSources(srcs []Source) error {
	for _, src := range srcs {
		d, err := src.LoadData()
		if err != nil {
			return err
		}

		opts, err := m.loadContent(src.GetType(), d)
		if err != nil {
			return err
		}
		m.options.Merge(opts)
	}
	return nil
}

func (m *Manager) loadContent(ct string, cd []byte) (data.Map, error) {
	opts := data.Map{}
	switch strings.ToLower(ct) {
	case "yaml", "yml":
		if err := yaml.Unmarshal(cd, &opts); err != nil {
			return nil, m.loadError(err)
		}
	case "json":
		if err := json.Unmarshal(cd, &opts); err != nil {
			return nil, m.loadError(err)
		}
		//case "xml":
		//	if err := xml.Unmarshal(cd, &opts); err != nil {
		//		return nil, loadError(err)
		//	}
	default:
		return nil, errors.New("unsupported config type: " + ct)
	}
	return opts, nil
}

func (m *Manager) loadError(err error) error {
	return errors.Wrap(err, "loading config failed")
}

// Get searches option from flag/env/config/remote/default. It returns nil if option is not found.
func (m *Manager) Get(key string) interface{} {
	// ensure loaded
	if !m.loaded {
		err := m.load(false)
		if err != nil {
			panic(err)
		}
	}

	opt := m.options.Find(key)
	def := m.defaults.Get(key)
	if def == nil {
		return opt
	} else if opt == nil {
		return def
	} else if v, err := cast.TryToValue(opt, reflect.TypeOf(def)); err == nil {
		return v
	}
	return opt
}

// GetInt returns option as int.
func (m *Manager) GetInt(key string) int {
	return cast.ToInt(m.Get(key))
}

// GetInt8 returns option as int8.
func (m *Manager) GetInt8(key string) int8 {
	return cast.ToInt8(m.Get(key))
}

// GetInt16 returns option as int16.
func (m *Manager) GetInt16(key string) int16 {
	return cast.ToInt16(m.Get(key))
}

// GetInt32 returns option as int32.
func (m *Manager) GetInt32(key string) int32 {
	return cast.ToInt32(m.Get(key))
}

// GetInt64 returns option as int64.
func (m *Manager) GetInt64(key string) int64 {
	return cast.ToInt64(m.Get(key))
}

// GetInt returns option as uint.
func (m *Manager) GetUint(key string) uint {
	return cast.ToUint(m.Get(key))
}

// GetInt8 returns option as uint8.
func (m *Manager) GetUint8(key string) uint8 {
	return cast.ToUint8(m.Get(key))
}

// GetInt16 returns option as uint16.
func (m *Manager) GetUint16(key string) uint16 {
	return cast.ToUint16(m.Get(key))
}

// GetInt32 returns option as uint32.
func (m *Manager) GetUint32(key string) uint32 {
	return cast.ToUint32(m.Get(key))
}

// GetInt64 returns option as uint64.
func (m *Manager) GetUint64(key string) uint64 {
	return cast.ToUint64(m.Get(key))
}

// GetString returns option as string.
func (m *Manager) GetString(key string) string {
	return cast.ToString(m.Get(key))
}

// GetBool returns option as bool.
func (m *Manager) GetBool(key string) bool {
	return cast.ToBool(m.Get(key))
}

// GetDuration returns option as time.Duration.
func (m *Manager) GetDuration(key string) time.Duration {
	return cast.ToDuration(m.Get(key))
}

// GetTime returns option as time.Time.
func (m *Manager) GetTime(key string) time.Time {
	return cast.ToTime(m.Get(key))
}
