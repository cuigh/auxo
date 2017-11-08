package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/net/web"
)

// FormValidator defines a function to validate FormAuth credentials.
type FormValidator func(ctx web.Context) (ticket string, err error)

// FormIdentifier defines a function to identify user ticket.
type FormIdentifier func(ticket string) web.User

// Form implements form authentication.
type Form struct {
	CookieName        string // default '_u'
	CookieDomain      string
	CookiePath        string
	DefaultUrl        string
	Remember          bool
	SlidingExpiration bool
	Timeout           time.Duration
	Identifier        FormIdentifier
}

// NewForm returns an Form authenticate filter.
func NewForm(identifier FormIdentifier) *Form {
	if identifier == nil {
		panic("identifier func can't be nil")
	}

	return &Form{
		CookieName: "_u",
		CookiePath: "/",
		DefaultUrl: "/",
		Identifier: identifier,
	}
}

func (f *Form) LoginForm(validator func(name, pwd string) (ticket string, err error)) web.HandlerFunc {
	return f.Login(func(ctx web.Context) (ticket string, err error) {
		name := ctx.F("name")
		pwd := ctx.F("password")
		return validator(name, pwd)
	})
}

func (f *Form) LoginJSON(validator func(name, pwd string) (ticket string, err error)) web.HandlerFunc {
	return f.Login(func(ctx web.Context) (ticket string, err error) {
		m := &struct {
			Name     string `json:"name"`
			Password string `json:"password"`
		}{}
		err = ctx.Bind(m)
		if err != nil {
			return
		}
		return validator(m.Name, m.Password)
	})
}

//func (f *Form) LoginFormCaptcha(validator func(name, pwd, captchaId, captchaCode string) (ticket string, err error)) web.HandlerFunc {
//	return f.Login(func(ctx web.Context) (ticket string, err error) {
//		name := ctx.F("name")
//		pwd := ctx.F("password")
//		id := ctx.F("captcha_id")
//		code := ctx.F("captcha_code")
//		return validator(name, pwd, id, code)
//	})
//}

// Login returns a handler for sign-in.
func (f *Form) Login(validator FormValidator) web.HandlerFunc {
	return func(ctx web.Context) error {
		ticket, err := validator(ctx)
		if err != nil {
			return err
		}

		f.renewTicket(ctx, ticket)

		url := ctx.Q("from")
		if url == "" {
			url = f.DefaultUrl
		}

		ct := ctx.Request().Header.Get(web.HeaderContentType)
		if strings.HasPrefix(ct, web.MIMEApplicationJSON) {
			return ctx.JSON(data.Map{
				"success": true,
				"url":     url,
			})
		}
		return ctx.Redirect(url)
	}
}

// Logout is a handler for sign-out.
func (f *Form) Logout(ctx web.Context) error {
	c := &http.Cookie{
		Name:   f.CookieName,
		MaxAge: -1,
	}
	ctx.SetCookie(c)
	return ctx.Redirect(f.DefaultUrl)
}

// Apply implements `web.Filter` interface.
func (f *Form) Apply(next web.HandlerFunc) web.HandlerFunc {
	if f.CookieName == "" {
		f.CookieName = "_u"
	}
	if f.CookiePath == "" {
		f.CookiePath = "/"
	}
	if f.DefaultUrl == "" {
		f.DefaultUrl = "/"
	}

	//logger := log.Get(PkgName)
	return func(ctx web.Context) error {
		//logger.Debug("form-auth:", ctx.Request().RequestURI)
		cookie, _ := ctx.Cookie(f.CookieName)
		if cookie != nil {
			user := f.Identifier(cookie.Value)
			ctx.SetUser(user)

			if f.SlidingExpiration && !f.Remember {
				f.renewTicket(ctx, cookie.Value)
			}
		}
		//logger.Debug("user:", ctx.User())

		return next(ctx)
	}
}

func (f *Form) renewTicket(ctx web.Context, ticket string) {
	c := &http.Cookie{
		Name:  f.CookieName,
		Value: ticket,
		Path:  f.CookiePath,
	}
	if f.Timeout > 0 {
		c.Expires = time.Now().Add(f.Timeout)
	}
	ctx.SetCookie(c)
}
