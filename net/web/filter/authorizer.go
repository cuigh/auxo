package filter

import (
	"net/http"
	"net/url"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/net/web"
)

type Authorizer struct {
	Checker         func(user web.User, handler web.HandlerInfo) bool
	LoginUrl        string
	UnauthorizedMsg string
	ForbiddenMsg    string
}

func NewAuthorizer(checker func(user web.User, handler web.HandlerInfo) bool) *Authorizer {
	return &Authorizer{
		Checker: checker,
	}
}

// Apply implements `web.Filter` interface.
func (a *Authorizer) Apply(next web.HandlerFunc) web.HandlerFunc {
	if a.Checker == nil {
		panic("Authorizer requires an checker function")
	}

	if a.LoginUrl == "" {
		a.LoginUrl = "/login"
	}
	if a.UnauthorizedMsg == "" {
		a.UnauthorizedMsg = "You are not logged in"
	}
	if a.ForbiddenMsg == "" {
		a.ForbiddenMsg = "You do not have access to this URL"
	}

	return func(ctx web.Context) error {
		auth := ctx.Handler().Authorize()
		if auth == web.AuthAnonymous {
			return next(ctx)
		}

		user := ctx.User()
		if user == nil || user.Anonymous() {
			ct := ctx.ContentType()
			if ct == web.MIMEApplicationJSON {
				//return web.NewError(http.StatusUnauthorized)
				return ctx.Status(http.StatusUnauthorized).JSON(data.Map{
					"url":       ctx.Route(),
					"code":      http.StatusUnauthorized,
					"message":   a.UnauthorizedMsg,
					"login_url": a.LoginUrl,
				})
			} else if ctx.IsAJAX() {
				return ctx.Status(http.StatusUnauthorized).HTML(a.UnauthorizedMsg)
			}

			u, err := url.Parse(a.LoginUrl)
			if err != nil {
				return err
			}
			q := u.Query()
			q.Set("from", ctx.Request().RequestURI)
			u.RawQuery = q.Encode()
			return ctx.Redirect(u.String())
		}

		if auth == web.AuthExplicit && !a.Checker(user, ctx.Handler()) {
			return web.NewError(http.StatusForbidden, a.ForbiddenMsg)
		}
		return next(ctx)
	}
}
