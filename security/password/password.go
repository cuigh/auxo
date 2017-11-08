package password

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"hash"
)

var (
	hasher    Hasher
	saltBytes = 32 // 256 bits
)

type Hasher func(pwd, salt string) (string, error)

type Manager struct {
	hasher Hasher
}

func NewManager(hasher Hasher, saltBytes int) *Manager {
	return nil
}

func (m *Manager) Password(pwd string) (p string, s string, e error) {
	return
}

func (m *Manager) Salt() (string, error) {
	b := make([]byte, saltBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (m *Manager) Validate(pwd, hash, salt string) bool {
	h, err := Hash(pwd, salt)
	return err == nil && h == hash
}

func SetHasher(h Hasher) {
	if h == nil {
		panic("hasher can't be nil")
	}
	hasher = h
}

func SetSaltBytes(bytes int) {
	if bytes <= 0 {
		panic("bytes <= 0")
	}
	saltBytes = bytes
}

func Validate(hashed, raw, salt string) bool {
	h, err := Hash(raw, salt)
	return err == nil && h == hashed
}

func Hash(pwd, salt string) (string, error) {
	return hasher(pwd, salt)
}

func Get(pwd string) (hash, salt string, err error) {
	salt, err = NewSalt()
	if err == nil {
		hash, err = Hash(pwd, salt)
	}
	return
}

func NewSalt() (string, error) {
	b := make([]byte, saltBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func PBKDF2() Hasher {
	return func(pwd, salt string) (string, error) {
		b := pbkdf2([]byte(pwd), []byte(salt), 4096, 32, sha1.New)
		return hex.EncodeToString(b), nil
	}
}

// not safe enough, but old system may use this
//func SHA256() Hasher {
//	return func(pwd, salt string) (string, error) {
//		h := sha256.New()
//		if _, err := h.Write([]byte(pwd + salt)); err != nil {
//			return "", err
//		}
//		b := h.Sum(nil)
//		return hex.EncodeToString(b), nil
//	}
//}

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

func init() {
	// not safe enough, but old system may use this
	//hasher = func(pwd, salt string) (string, error) {
	//	h := sha256.New()
	//	if _, err := h.Write([]byte(pwd + salt)); err != nil {
	//		return "", err
	//	}
	//	b := h.Sum(nil)
	//	return hex.EncodeToString(b), nil
	//}
	hasher = func(pwd, salt string) (string, error) {
		b := pbkdf2([]byte(pwd), []byte(salt), 4096, 32, sha1.New)
		return hex.EncodeToString(b), nil
	}
}
