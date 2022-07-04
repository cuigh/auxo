package ioc

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

type User struct {
	Id   int32
	Name string
}

type UserService struct {
}

func (u *UserService) Find(id int32) *User {
	return &User{Id: id, Name: "Auxo"}
}

type AuthService struct {
	User *UserService
}

func (u *AuthService) GetName(id int32) string {
	return u.User.Find(id).Name
}

func TestContainer_PutFind(t *testing.T) {
	c := New()

	c.Put(func() *UserService { return &UserService{} }, Name("User"))
	c.Put(func(u *UserService) *AuthService {
		return &AuthService{u}
	}, Name("Auth"))

	assert.NotNil(t, c.Find("User"))

	auth, err := c.TryFind("Auth")
	assert.NoError(t, err)
	assert.NotNil(t, auth)

	test, err := c.TryFind("Test")
	assert.Error(t, err)
	assert.Nil(t, test)
}

func TestContainer_Call(t *testing.T) {
	c := New()
	c.Put(func() *UserService { return &UserService{} }, Name("User"))
	c.Put(func(u *UserService) *AuthService {
		return &AuthService{u}
	}, Name("Auth"))

	err := c.Call(func(s *AuthService) {
		t.Log(s.GetName(1))
	})
	assert.NoError(t, err)

	err = c.Call(func(u *User) {
		t.Log(u.Name)
	})
	assert.Error(t, err)
}

func TestBind(t *testing.T) {
	c := New()
	c.Put(func() *User {
		return &User{1, "noname"}
	}, Name("User"))

	target := &struct {
		User  *User `container:"name"`
		User1 *User `container:"type"`
	}{}
	err := c.Bind(target)

	assert.NoError(t, err)
	assert.NotNil(t, target.User)
	assert.NotNil(t, target.User1)
}
