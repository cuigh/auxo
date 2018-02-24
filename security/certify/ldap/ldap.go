package ldap

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"

	"github.com/cuigh/auxo/log"
	"github.com/cuigh/auxo/security"
	"github.com/cuigh/auxo/security/certify"
	"github.com/go-ldap/ldap"
)

const PkgName = "auxo.security.certify.ldap"

type User struct {
	id        string
	loginName string
	name      string
	email     string
}

func (u *User) ID() string {
	return u.id
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Anonymous() bool {
	return u.id == ""
}

func (u *User) LoginName() string {
	return u.loginName
}

func (u *User) Email() string {
	return u.email
}

type SecurityPolicy int32

const (
	SecurityNone     = SecurityPolicy(0)
	SecurityTLS      = SecurityPolicy(1)
	SecurityStartTLS = SecurityPolicy(2)
)

type Option func(*Realm)

func Security(security SecurityPolicy) Option {
	return func(r *Realm) {
		r.security = security
	}
}

func UserFilter(filter string) Option {
	return func(r *Realm) {
		if filter != "" {
			r.userFilter = filter
		}
	}
}

// Binding enables binding authentication
func Binding(dn, pwd string) Option {
	return func(r *Realm) {
		r.bindDN = dn
		r.bindPwd = pwd
	}
}

func NameAttr(attr string) Option {
	return func(r *Realm) {
		if attr != "" {
			r.nameAttr = attr
		}
	}
}

func EmailAttr(attr string) Option {
	return func(r *Realm) {
		if attr != "" {
			r.emailAttr = attr
		}
	}
}

type Realm struct {
	logger log.Logger

	// options
	addr     string
	security SecurityPolicy
	//auth       AuthPolicy
	bindDN     string // DN to bind with
	bindPwd    string // Bind DN password
	baseDN     string // Base search path for users
	userDN     string // Template for the DN of the user for simple auth
	userFilter string // Search filter for user
	nameAttr   string
	emailAttr  string
	//tlsCert        string // TLS cert
	//tlsVerify      bool   // Verify cert
}

func New(addr, baseDN, userDN string, opts ...Option) certify.Realm {
	r := &Realm{
		addr:       addr,
		baseDN:     baseDN,
		userDN:     userDN,
		userFilter: "(&(objectClass=user)(sAMAccountName=%s))",
		nameAttr:   "displayName",
		emailAttr:  "mail",
		logger:     log.Get(PkgName),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (*Realm) Name() string {
	return "ldap"
}

func (r *Realm) Login(token certify.Token) (security.User, error) {
	st, ok := token.(*certify.SimpleToken)
	if !ok {
		return nil, certify.ErrInvalidToken
	}

	c, err := r.dial()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// bind
	err = r.bind(c, st.Name(), st.Password())
	if err != nil {
		r.logger.Error("ldap > Failed to bind: ", err)
		return nil, certify.ErrInvalidToken
	}

	// If user wasn't exist, we need create it
	entry, err := r.search(c, st.Name(), r.nameAttr, r.emailAttr)
	if err != nil {
		return nil, err
	}

	return &User{
		loginName: st.Name(),
		name:      entry.GetAttributeValue(r.nameAttr),
		email:     entry.GetAttributeValue(r.emailAttr),
	}, nil
}

func (r *Realm) bind(c *ldap.Conn, name, pwd string) (err error) {
	if r.bindDN == "" {
		// simple auth
		err = c.Bind(fmt.Sprintf(r.userDN, name), pwd)
	} else {
		// bind auth
		err = c.Bind(r.bindDN, r.bindPwd)
		if err == nil {
			var entry *ldap.Entry
			entry, err = r.search(c, name, "cn")
			if err == nil {
				err = c.Bind(entry.DN, pwd)
			}
		}
	}
	return
}

func (r *Realm) search(c *ldap.Conn, name string, attrs ...string) (entry *ldap.Entry, err error) {
	filter := fmt.Sprintf(r.userFilter, name)
	req := ldap.NewSearchRequest(
		r.baseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter, attrs, nil,
	)
	sr, err := c.Search(req)
	if err != nil {
		return nil, err
	}

	if length := len(sr.Entries); length == 0 {
		return nil, errors.New("ldap: user not found with filter: " + filter)
	} else if length > 1 {
		return nil, errors.New("ldap: found more than one user when searching with filter: " + filter)
	}

	return sr.Entries[0], nil
}

func (r *Realm) dial() (*ldap.Conn, error) {
	if r.security == SecurityNone {
		return ldap.Dial("tcp", r.addr)
	}

	host, _, err := net.SplitHostPort(r.addr)
	if err != nil {
		return nil, err
	}

	// TODO: support tls cert and verification
	tc := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
		Certificates:       nil,
	}

	if r.security == SecurityTLS {
		return ldap.DialTLS("tcp", r.addr, tc)
	}

	conn, err := ldap.Dial("tcp", r.addr)
	if err != nil {
		return nil, err
	}

	if err = conn.StartTLS(tc); err != nil {
		conn.Close()
		log.Get("user").Error("ldap > Failed to switch to TLS: ", err)
		return nil, err
	}
	return conn, nil
}
