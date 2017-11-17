package valid_test

import (
	"testing"

	"github.com/cuigh/auxo/data/valid"
	"github.com/cuigh/auxo/test/assert"
)

func TestRequired(t *testing.T) {
	cases := []struct {
		Name string `valid:"required"`
		OK   bool
	}{
		{"xyz", true},
		{"", false},
	}

	v := &valid.Validator{}
	for _, c := range cases {
		err := v.Validate(c)
		if c.OK {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestEmail(t *testing.T) {
	cases := []struct {
		Email string `valid:"email"`
		OK    bool
	}{
		{"xyz", false},
		{"test@test.com", true},
		{"test.a_b-c.1@test.com", true},
	}

	v := &valid.Validator{}
	for _, c := range cases {
		err := v.Validate(c)
		if c.OK {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestURL(t *testing.T) {
	cases := []struct {
		URL string `valid:"url"`
		OK  bool
	}{
		{"xyz", false},
		{"http://google.com", true},
		{"http://www.google.com", true},
		{"https://www.google.com/", true},
		{"https://www.google.com/abc", true},
	}

	v := &valid.Validator{}
	for _, c := range cases {
		err := v.Validate(c)
		if c.OK {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestAlpha(t *testing.T) {
	cases := []struct {
		URL string `valid:"alpha"`
		OK  bool
	}{
		{"xyz", true},
		{"123", false},
	}

	v := &valid.Validator{}
	for _, c := range cases {
		err := v.Validate(c)
		if c.OK {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestLength(t *testing.T) {
	cases := []struct {
		Name string `valid:"length[5~)"`
		OK   bool
	}{
		{"test", false},
		{"abcde", true},
	}

	v := &valid.Validator{}
	for _, c := range cases {
		err := v.Validate(c)
		if c.OK {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestRange(t *testing.T) {
	cases := []struct {
		Value int `valid:"range[5~10)"`
		OK    bool
	}{
		{1, false},
		{11, false},
		{7, true},
	}

	v := &valid.Validator{}
	for _, c := range cases {
		err := v.Validate(c)
		if c.OK {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestIPV4(t *testing.T) {
	cases := []struct {
		IP string `valid:"ipv4"`
		OK bool
	}{
		{"192.0.2.1", true},
		{"2001:db8::68", false},
	}

	v := &valid.Validator{}
	for _, c := range cases {
		err := v.Validate(c)
		if c.OK {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestIPV6(t *testing.T) {
	cases := []struct {
		IP string `valid:"ipv6"`
		OK bool
	}{
		{"192.0.2.1", false},
		{"2001:db8::68", true},
	}

	v := &valid.Validator{}
	for _, c := range cases {
		err := v.Validate(c)
		if c.OK {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestRegex(t *testing.T) {
	cases := []struct {
		Number string `valid:"regex(^\\d+$)"`
		OK     bool
	}{
		{"123", true},
		{"abc", false},
	}

	v := &valid.Validator{}
	for _, c := range cases {
		err := v.Validate(c)
		if c.OK {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
