package validator

import "github.com/cuigh/auxo/data/valid"

type Validator valid.Validate

func (v *Validator) Validate(i interface{}) error {
	return (*valid.Validate)(v).Struct(i)
}

func New() *Validator {
	return (*Validator)(valid.New())
}
