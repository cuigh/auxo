package passwd

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"hash"
)

var Default = &Manager{
	Hasher:    PBKDF2(4096),
	SaltBytes: 32,
}

type Hasher func(pwd, salt string) (string, error)

type Manager struct {
	Hasher
	SaltBytes int
}

// Validate validates the password is correct.
func (m *Manager) Validate(pwd, hash, salt string) bool {
	h, err := m.Hasher(pwd, salt)
	return err == nil && h == hash
}

// Generate create a hashed password with random salt.
func (m *Manager) Generate(raw string) (pwd string, salt string, err error) {
	salt, err = m.createSalt()
	if err == nil {
		pwd, err = m.Hasher(raw, salt)
	}
	return
}

func (m *Manager) createSalt() (string, error) {
	b := make([]byte, m.SaltBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Validate validates the password is correct.
func Validate(pwd, hash, salt string) bool {
	return Default.Validate(pwd, hash, salt)
}

// Generate create a hashed password with random salt.
func Generate(raw string) (pwd string, salt string, err error) {
	return Default.Generate(raw)
}

// PBKDF2 is a Hasher implement with PBKDF2 algorithm.
func PBKDF2(iter int) Hasher {
	return func(pwd, salt string) (string, error) {
		b := pbkdf2([]byte(pwd), []byte(salt), iter, 32, sha1.New)
		return hex.EncodeToString(b), nil
	}
}

// SHA256 is a Hasher implement with SHA256 algorithm.
func SHA256(pwd, salt string) (string, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(pwd + salt)); err != nil {
		return "", err
	}
	b := h.Sum(nil)
	return hex.EncodeToString(b), nil
}

// copy from: golang.org/x/crypto/pbkdf2
func pbkdf2(password, salt []byte, iter, keyLen int, h func() hash.Hash) []byte {
	prf := hmac.New(h, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var buf [4]byte
	dk := make([]byte, 0, numBlocks*hashLen)
	U := make([]byte, hashLen)
	for block := 1; block <= numBlocks; block++ {
		// N.B.: || means concatenation, ^ means XOR
		// for each block T_i = U_1 ^ U_2 ^ ... ^ U_iter
		// U_1 = PRF(password, salt || uint(i))
		prf.Reset()
		prf.Write(salt)
		buf[0] = byte(block >> 24)
		buf[1] = byte(block >> 16)
		buf[2] = byte(block >> 8)
		buf[3] = byte(block)
		prf.Write(buf[:4])
		dk = prf.Sum(dk)
		T := dk[len(dk)-hashLen:]
		copy(U, T)

		// U_n = PRF(password, U_(n-1))
		for n := 2; n <= iter; n++ {
			prf.Reset()
			prf.Write(U)
			U = U[:0]
			U = prf.Sum(U)
			for x := range U {
				T[x] ^= U[x]
			}
		}
	}
	return dk[:keyLen]
}
