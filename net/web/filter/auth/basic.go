package auth

import (
	"encoding/base64"

	"github.com/cuigh/auxo/net/web"
)

// Basic implements Basic authentication.
type Basic struct {
	Validator func(name string, pwd string) web.User
}

// Apply implements `web.Filter` interface.
func (b *Basic) Apply(next web.HandlerFunc) web.HandlerFunc {
	const name = "Basic"

	if b.Validator == nil {
		panic("basic-auth requires a validator function")
	}

	return func(c web.Context) error {
		auth := c.Header(web.HeaderAuthorization)
		l := len(name)

		if len(auth) > l+1 && auth[:l] == name {
			buf, err := base64.StdEncoding.DecodeString(auth[l+1:])
			if err != nil {
				return err
			}
			cred := string(buf)
			for i := 0; i < len(cred); i++ {
				if cred[i] == ':' {
					if user := b.Validator(cred[:i], cred[i+1:]); user != nil {
						c.SetUser(user)
						return next(c)
					}
				}
			}
		}

		// Need to return `401` for browsers to pop-up login box.
		c.Response().Header().Set(web.HeaderWWWAuthenticate, name+" realm=Restricted")
		return web.ErrUnauthorized
	}
}

// NewBasic returns an Basic authenticate filter.
func NewBasic(validator func(name string, pwd string) web.User) *Basic {
	return &Basic{
		Validator: validator,
	}
}
