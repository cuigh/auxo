package passwd

import (
	"crypto/sha512"
	"encoding/hex"
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestHash(t *testing.T) {
	pwd := "abcde"
	p1, _, err := Generate(pwd)
	assert.NoError(t, err)
	p2, _, err := Generate(pwd)
	assert.NoError(t, err)
	assert.NotEqual(t, p1, p2)
}

func TestValidate(t *testing.T) {
	pwd := "abcde"

	hash, salt, err := Generate(pwd)
	assert.NoError(t, err)

	ok := Validate(pwd, hash, salt)
	assert.True(t, ok)
}

func TestSetHasher(t *testing.T) {
	hasher := func(pwd, salt string) (string, error) {
		h := sha512.New()
		if _, err := h.Write([]byte(pwd + salt)); err != nil {
			return "", err
		}
		b := h.Sum(nil)
		return hex.EncodeToString(b), nil
	}
	m := &Manager{
		Hasher:    hasher,
		SaltBytes: 32,
	}
	h, _, err := m.Generate("1")
	assert.NoError(t, err)
	assert.Equal(t, (512/8)*2, len(h))
}

func TestGenerate(t *testing.T) {
	p, s, err := Generate("123456")
	assert.NoError(t, err)
	t.Log(p, s)
}

func TestSetSaltBytes(t *testing.T) {
	hasher := func(pwd, salt string) (string, error) {
		h := sha512.New()
		if _, err := h.Write([]byte(pwd + salt)); err != nil {
			return "", err
		}
		b := h.Sum(nil)
		return hex.EncodeToString(b), nil
	}
	m := &Manager{
		Hasher:    hasher,
		SaltBytes: 16,
	}
	_, salt, err := m.Generate("1")
	assert.NoError(t, err)
	assert.Equal(t, 16*2, len(salt))
}
