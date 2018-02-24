package certify

import (
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/security"
)

var (
	ErrInvalidToken      = errors.New("security: invalid principal or credential")
	ErrInvalidPrincipal  = errors.New("security: principal is invalid")
	ErrInvalidCredential = errors.New("security: credential is invalid")
	ErrAccountLocked     = errors.New("security: account is locked")
	ErrAccountDisabled   = errors.New("security: account is disabled")
)

type Token interface {
	Principal() interface{}
	Credential() interface{}
}

type SimpleToken struct {
	name    string
	pwd     string
	captcha string
}

func NewSimpleToken(name, pwd string, captcha ...string) Token {
	t := &SimpleToken{
		name: name,
		pwd:  pwd,
	}
	if len(captcha) > 0 {
		t.captcha = captcha[0]
	}
	return t
}

func (t *SimpleToken) Principal() interface{} {
	return t.name
}

func (t *SimpleToken) Credential() interface{} {
	return t.pwd
}

func (t *SimpleToken) Name() string {
	return t.name
}

func (t *SimpleToken) Password() string {
	return t.pwd
}

func (t *SimpleToken) Captcha() string {
	return t.captcha
}

type Realm interface {
	// Name is the name of realm.
	Name() string
	// Login does authentication with token, it returns an error if failed, or nil user if skipped.
	Login(token Token) (security.User, error)
}

//type AuthenticatePolicy func(realms []Realm, token Token) (User, error)

type Authenticator struct {
	realms []Realm
}

func (a *Authenticator) AddRealm(r ...Realm) {
	a.realms = append(a.realms, r...)
}

func (a *Authenticator) Login(token Token) (realm string, user security.User, err error) {
	for _, r := range a.realms {
		user, err = r.Login(token)
		if err != nil {
			continue
		}

		if user != nil {
			return r.Name(), user, nil
		}
	}

	if err == nil {
		err = ErrInvalidToken
	}
	return
}

type simpleRealm struct {
	name  string
	login func(token *SimpleToken) (security.User, error)
}

func NewSimpleRealm(name string, login func(token *SimpleToken) (security.User, error)) Realm {
	return &simpleRealm{
		name:  name,
		login: login,
	}
}

func (r *simpleRealm) Name() string {
	return r.name
}

func (r *simpleRealm) Login(token Token) (security.User, error) {
	if st, ok := token.(*SimpleToken); ok {
		return r.login(st)
	}
	return nil, ErrInvalidToken
}
