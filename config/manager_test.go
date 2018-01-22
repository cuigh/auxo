package config_test

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cuigh/auxo/byte/size"
	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/test/assert"
)

func initManager() *config.Manager {
	m := config.New("app")
	m.AddFolder("./samples")
	return m
}

func TestFindFile(t *testing.T) {
	m := initManager()
	f := m.FindFile("app", ".yml")
	assert.NotEmpty(t, f)
}

func TestFindFiles(t *testing.T) {
	m := initManager()
	fs := m.FindFiles("app", ".json", ".yml")
	assert.True(t, len(fs) == 2)
}

//func TestFindFolder(t *testing.T) {
//	m := initManager()
//	f := m.FindFolder("test")
//	assert.NotEmpty(t, f)
//}

//func TestFindFolders(t *testing.T) {
//	m := initManager()
//	fs := m.FindFolders("etcd2")
//	assert.True(t, len(fs) == 1)
//}

func TestType(t *testing.T) {
	cases := []struct {
		Key      string
		Expected string
	}{
		{`yaml.name`, "yaml"},
		{`json.name`, "json"},
	}
	m := initManager()

	for _, c := range cases {
		assert.Equal(t, c.Expected, m.Get(c.Key))
	}
}

func TestGet(t *testing.T) {
	m := initManager()
	m.AddFolder(".")
	m.SetProfile("dev")
	assert.Equal(t, int(10), m.GetInt("test.id"))
	assert.Equal(t, int8(10), m.GetInt8("test.id"))
	assert.Equal(t, int16(10), m.GetInt16("test.id"))
	assert.Equal(t, int32(10), m.GetInt32("test.id"))
	assert.Equal(t, int64(10), m.GetInt64("test.id"))
}

func TestProfile(t *testing.T) {
	source := `name: auxo`
	m := initManager()
	m.AddDataSource([]byte(source), "yaml")
	assert.Equal(t, "auxo", m.Get("name"))
}

func TestFlags(t *testing.T) {
	m := initManager()

	fs := flag.NewFlagSet("", flag.PanicOnError)
	fs.String("version", "1.0", "app version")
	fs.String("a.b.c", "test", "app version")
	m.BindFlags(fs)

	version := m.Get("version")
	assert.Equal(t, "1.0", version)

	c := m.Get("a.b")
	assert.Equal(t, "map[c:test]", fmt.Sprint(c))
}

func TestEnv(t *testing.T) {
	const (
		dbAddressKey    = "db.address"
		dbAddressEnvKey = "DB_ADDRESS"
		dbAddress       = "127.0.0.1"
	)

	os.Setenv(dbAddressEnvKey, dbAddress)
	os.Setenv("A_B_C", "test")

	m := initManager()
	m.SetEnvPrefix("")
	m.BindEnv(dbAddressKey, dbAddressEnvKey)

	cases := []struct {
		key   string
		value string
	}{
		{dbAddressKey, dbAddress},
		{"a.b.c", "test"},
	}

	for _, c := range cases {
		value := m.Get(c.key)
		assert.Equal(t, c.value, value)
	}
}

func TestDefaultValue(t *testing.T) {
	m := initManager()
	m.SetDefaultValue("default.value", 1)
	v := m.Get("default.value")
	assert.Equal(t, 1, v)
}

func TestConfigUnmarshal(t *testing.T) {
	source := `
name: 'auxo'
debug: true
`
	v := struct {
		Name  string
		Debug bool
	}{}
	m := initManager()
	m.AddDataSource([]byte(source), "yaml")

	err := m.Unmarshal(&v)
	assert.NoError(t, err)
	t.Log(v)
}

func TestConfigUnmarshalWeb(t *testing.T) {
	type WebConfig struct {
		Debug             bool
		RunMode           string        `option:"mode"`    // dev/prd
		Addresses         []string      `option:"address"` // http=:8001,https=:443,unix=/a/b
		ReadTimeout       time.Duration // `option:"read_timeout"`
		ReadHeaderTimeout time.Duration `option:"read_header_timeout"`
		WriteTimeout      time.Duration `option:"write_timeout"`
		IdleTimeout       time.Duration `option:"idle_timeout"`
		MaxHeaderBytes    int
		IndexUrl          string
		LoginUrl          string
		UnauthorizedUrl   string
		Time              MyTime
		MaxBodySize       size.Size
		//DisablePathCorrection bool
		//Compression bool
		//SessionCookieName string // auxo.web
		ACME struct {
			Dir   string
			Email string
		}
	}
	m := initManager()

	v := WebConfig{}
	err := m.UnmarshalOption("web", &v)
	assert.NoError(t, err)
	t.Logf("%+v", v)
}

type MyTime time.Time

func (t *MyTime) String() string {
	return time.Time(*t).String()
}

func (t *MyTime) Unmarshal(i interface{}) error {
	s := i.(string)
	if s == "s" {
		*t = MyTime(time.Now())
	}

	time.Now().MarshalJSON()
	return nil
}

func TestConfigEnv(t *testing.T) {
	const value = "auxo"
	os.Setenv("TEST_NAME", value)
	m := initManager()
	assert.Equal(t, value, m.Get("test.name"))
}
