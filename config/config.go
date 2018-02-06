// Package config manages configurations of application.
package config

import (
	"flag"
	"time"
)

var (
	// default manager
	m = New("app")
)

// FindFile searches all config directories and return the first found file.
func FindFile(name string, exts ...string) string {
	return m.FindFile(name, exts...)
}

// FindFile searches all config directories and return all found files.
func FindFiles(name string, exts ...string) []string {
	return m.FindFiles(name, exts...)
}

// FindFolder searches all config directories and return the first found folder.
func FindFolder(name string) string {
	return m.FindFolder(name)
}

// FindFolders searches all config directories and return all found folders.
func FindFolders(name string) []string {
	return m.FindFolders(name)
}

// BindFlags binds a flag set.
func BindFlags(set *flag.FlagSet) {
	m.BindFlags(set)
}

// SetEnvPrefix sets the prefix of environment variables. Default prefix is "AUXO".
func SetEnvPrefix(prefix string) {
	m.SetEnvPrefix(prefix)
}

// BindEnv binds a option to an environment variable.
func BindEnv(key string, envKey string) {
	m.BindEnv(key, envKey)
}

// SetProfiles sets active profiles. Profiles are only valid to local file sources.
func SetProfile(profiles ...string) {
	m.SetProfile(profiles...)
}

// SetName sets name of the main configuration file (without extension).
func SetName(name string) {
	m.SetName(name)
}

// AddFolders adds directories to searching list.
func AddFolder(dirs ...string) {
	m.AddFolder(dirs...)
}

// AddSource adds custom configuration sources.
func AddSource(srcs ...Source) {
	m.AddSource(srcs...)
}

// AddDataSource add a config source with bytes and type.
func AddDataSource(data []byte, ct string) {
	m.AddDataSource(data, ct)
}

// SetDefaultValue sets a default option.
func SetDefaultValue(name string, value interface{}) {
	m.SetDefaultValue(name, value)
}

// Load reads options from all sources.
func Load() error {
	return m.Load()
}

// Unmarshal exports options to struct.
func Unmarshal(v interface{}) error {
	return m.Unmarshal(v)
}

// Unmarshal exports specific option to struct.
func UnmarshalOption(name string, v interface{}) error {
	return m.UnmarshalOption(name, v)
}

// Exist returns the key is exist or not.
func Exist(key string) bool {
	return m.Get(key) != nil
}

// Get searches option from flag/env/config/remote/default. It returns nil if option is not found.
func Get(key string) interface{} {
	return m.Get(key)
}

// GetInt returns option as int.
func GetInt(key string) int {
	return m.GetInt(key)
}

// GetInt8 returns option as int8.
func GetInt8(key string) int8 {
	return m.GetInt8(key)
}

// GetInt16 returns option as int16.
func GetInt16(key string) int16 {
	return m.GetInt16(key)
}

// GetInt32 returns option as int32.
func GetInt32(key string) int32 {
	return m.GetInt32(key)
}

// GetInt64 returns option as int64.
func GetInt64(key string) int64 {
	return m.GetInt64(key)
}

// GetInt returns option as uint.
func GetUint(key string) uint {
	return m.GetUint(key)
}

// GetInt8 returns option as uint8.
func GetUint8(key string) uint8 {
	return m.GetUint8(key)
}

// GetInt16 returns option as uint16.
func GetUint16(key string) uint16 {
	return m.GetUint16(key)
}

// GetInt32 returns option as uint32.
func GetUint32(key string) uint32 {
	return m.GetUint32(key)
}

// GetInt64 returns option as uint64.
func GetUint64(key string) uint64 {
	return m.GetUint64(key)
}

// GetString returns option as string.
func GetString(key string) string {
	return m.GetString(key)
}

// GetBool returns option as bool.
func GetBool(key string) bool {
	return m.GetBool(key)
}

// GetDuration returns option as time.Duration.
func GetDuration(key string) time.Duration {
	return m.GetDuration(key)
}

// GetTime returns option as time.Time.
func GetTime(key string) time.Time {
	return m.GetTime(key)
}
