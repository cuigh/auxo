package passwd

import (
	"crypto/sha512"
	"encoding/hex"
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestHash(t *testing.T) {
	p1, err := Hash("abcde", "1")
	assert.NoError(t, err)

	p2, err := Hash("abcde", "2")
	assert.NoError(t, err)

	assert.NotEqual(t, p1, p2)
}

func TestValidate(t *testing.T) {
	pwd := "abcde"
	salt := "1"

	h, err := Hash(pwd, salt)
	assert.NoError(t, err)

	ok := Validate(h, pwd, salt)
	assert.True(t, ok)
}

func TestSetHasher(t *testing.T) {
	hasher = func(pwd, salt string) (string, error) {
		h := sha512.New()
		if _, err := h.Write([]byte(pwd + salt)); err != nil {
			return "", err
		}
		b := h.Sum(nil)
		return hex.EncodeToString(b), nil
	}
	SetHasher(hasher)

	h, err := Hash("1", "1")
	assert.NoError(t, err)
	assert.Equal(t, (512/8)*2, len(h))
}

func TestGet(t *testing.T) {
	p, s, err := Get("test")
	assert.NoError(t, err)
	t.Log(p, s)
}

func TestNewSalt(t *testing.T) {
	salt, err := NewSalt()
	assert.NoError(t, err)
	assert.Equal(t, saltBytes*2, len(salt))
}

func TestSetSaltBytes(t *testing.T) {
	SetSaltBytes(16)
	salt, err := NewSalt()
	assert.NoError(t, err)
	assert.Equal(t, 16*2, len(salt))
}
