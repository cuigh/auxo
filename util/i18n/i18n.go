package i18n

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/cuigh/auxo/config"
	"github.com/cuigh/auxo/encoding/yaml"
	"github.com/cuigh/auxo/ext/files"
)

var mgr = new(Manager)

// Get returns a Translator explicitly matched.
func Get(lang string) (Translator, error) {
	return mgr.Get(lang)
}

// Find returns any Translator according to language prefix.
func Find(langs ...string) (t Translator, err error) {
	for _, lang := range langs {
		t, err = mgr.Find(lang)
		if t != nil || err != nil {
			return
		}
	}
	return
}

// All returns all translators.
func All() (TranslatorMap, error) {
	return mgr.All()
}

type Manager struct {
	dirs []string
}

func New(dirs ...string) *Manager {
	return &Manager{
		dirs: dirs,
	}
}

// Find returns a translator by language prefix.
func (m *Manager) Find(lang string) (t Translator, err error) {
	// "zh-hans-cn" yields {"zh-hans-cn", "zh-hans", "zh"}
	t, err = m.Get(lang)
	if t != nil || err != nil {
		return
	}

	for i := strings.LastIndexByte(lang, '-'); i != -1; i = strings.LastIndexByte(lang, '-') {
		t, err = m.Get(lang[:i])
		if t != nil || err != nil {
			return
		}
	}
	return
}

// Get returns a translator by language.
func (m *Manager) Get(lang string) (t Translator, err error) {
	dirs := m.getDirs()
	for _, dir := range dirs {
		filename := filepath.Join(dir, lang+".yml")
		if files.Exist(filename) {
			return NewTranslator(lang, filename)
		}
	}
	return
}

// All returns all translators.
func (m *Manager) All() (TranslatorMap, error) {
	tm := TranslatorMap{}
	dirs := m.getDirs()
	for _, dir := range dirs {
		matches, err := filepath.Glob(filepath.Join(dir, "*.yml"))
		if err != nil {
			return nil, err
		}

		for _, f := range matches {
			lang := strings.TrimSuffix(filepath.Base(f), ".yml")
			t, err := NewTranslator(lang, f)
			if err != nil {
				return nil, err
			}
			tm[lang] = t
		}
	}
	return tm, nil
}

func (m *Manager) getDirs() []string {
	if len(m.dirs) == 0 {
		m.dirs = config.FindFolders("i18n")
	}
	return m.dirs
}

type TranslatorMap map[string]Translator

// Get returns a translator by language.
func (tm TranslatorMap) Get(lang string) Translator {
	return tm[lang]
}

// Find tries to find a translator by language prefix.
func (tm TranslatorMap) Find(lang string) Translator {
	// "zh-hans-cn": "zh-hans-cn" > "zh-hans" > "zh"
	if t := tm[lang]; t != nil {
		return t
	}

	for i := strings.LastIndexByte(lang, '-'); i != -1; i = strings.LastIndexByte(lang, '-') {
		if t := tm[lang[:i]]; t != nil {
			return t
		}
	}
	return nil
}

type Translator interface {
	Get(key string) string
	Format(key string, args ...interface{}) string
	Execute(key string, i interface{}) string // i must be map or struct
}

func NewTranslator(lang string, filename string) (Translator, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	t := &translator{
		lang: lang,
		m:    make(map[string]string),
	}
	err = yaml.Unmarshal(buf, &t.m)
	if err != nil {
		return nil, err
	}
	return t, nil
}

type translator struct {
	lang string
	m    map[string]string
}

func (t *translator) Get(key string) string {
	return t.m[key]
}

func (t *translator) Format(key string, args ...interface{}) string {
	if format := t.m[key]; format != "" {
		return fmt.Sprintf(format, args...)
	}
	return ""
}

func (t *translator) Execute(key string, i interface{}) string {
	if format := t.m[key]; format != "" {
		t, err := template.New("").Parse(format)
		if err != nil {
			panic(fmt.Sprintf("Parsing template failed, key: %v, data: %v, error: %v", key, i, err))
		}

		buf := &bytes.Buffer{}
		err = t.Execute(buf, i)
		if err != nil {
			panic(fmt.Sprintf("Executing template failed, key: %v, data: %v, error: %v", key, i, err))
		}
		return buf.String()
	}
	return ""
}

type combinedTranslator struct {
	ts []Translator
}

// Combine returns a merged Translator.
func Combine(ts ...Translator) (t Translator) {
	return &combinedTranslator{ts: ts}
}

func (ct *combinedTranslator) Get(key string) string {
	for _, t := range ct.ts {
		if msg := t.Get(key); msg != "" {
			return msg
		}
	}
	return ""
}

func (ct *combinedTranslator) Format(key string, args ...interface{}) string {
	for _, t := range ct.ts {
		if msg := t.Format(key, args...); msg != "" {
			return msg
		}
	}
	return ""
}

func (ct *combinedTranslator) Execute(key string, i interface{}) string {
	for _, t := range ct.ts {
		if msg := t.Execute(key, i); msg != "" {
			return msg
		}
	}
	return ""
}
