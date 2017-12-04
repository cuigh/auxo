package security

var Anonymous User = &user{}

type User interface {
	ID() string
	Name() string
	Anonymous() bool
	//Admin() bool
	//Roles() []string
	//Data(key string) interface{}
}

type user struct {
	id   string
	name string
}

func NewUser(id, name string) User {
	return &user{
		id:   id,
		name: name,
	}
}

func (u *user) ID() string {
	return u.id
}

func (u *user) Name() string {
	return u.name
}

func (u *user) Anonymous() bool {
	return u.id == ""
}
