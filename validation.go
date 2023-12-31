package valigator

import (
	"errors"
)

// Validator is an interface for any type that validates (the contents of) type T. It contains a single
// method, Validate, that returns an error if there is invalid or unexpectedly unset data in the input data structure.
type Validator[T any] interface {
	// Validate verifies if the input data structure contains invalid or missing data, returning an error if so.
	Validate(value T) error
}

// Func is a function type that complies with the Validator's Validate method signature.
//
// The Func type implements the Validator interface, through a Validate method calling on itself.
type Func[T any] func(T) error

// Validate implements the Validator interface.
//
// It verifies if the input data structure contains invalid or missing data, returning an error if so, by calling the
// inner Func with the input value.
func (fn Func[T]) Validate(value T) error {
	if fn == nil {
		return nil
	}

	return fn(value)
}

type multiValidator[T any] struct {
	validators []Validator[T]
}

// Validate implements the Validator interface.
//
// It verifies if the input data structure contains invalid or missing data, returning an error if so, by iterating
// through all configured Validator, while calling their Validate method on the input value.
func (v multiValidator[T]) Validate(value T) error {
	errs := make([]error, 0, len(v.validators))

	for i := range v.validators {
		if err := v.validators[i].Validate(value); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// New creates a Validator from the input slice of Func.
//
// If the input slice contains no items, this call returns a NoOp Validator. If it only contains one function, it
// will return it as a Func type, effectively as a Validator.
//
// If there are multiple functions in the input, a multi-Validator is created. This multi-Validator will contain all
// non-nil validators from the input, that will work with the same input value in one go.
func New[T any](validators ...func(T) error) Validator[T] {
	validators = nonNilFunc(validators)

	switch len(validators) {
	case 0:
		return NoOp[T]()
	case 1:
		return Func[T](validators[0])
	}

	mv := multiValidator[T]{
		validators: make([]Validator[T], 0, len(validators)),
	}

	for i := range validators {
		mv.validators = append(mv.validators, Func[T](validators[i]))
	}

	return mv
}

// Join gathers multiple Validator for the same type, joining them as a single Validator. It is similar to New,
// but works exclusively with Validator types as input.
func Join[T any](validators ...Validator[T]) Validator[T] {
	validators = nonNilValidators[T](validators)

	switch len(validators) {
	case 0:
		return NoOp[T]()
	case 1:
		return validators[0]
	}

	mv := &multiValidator[T]{
		validators: make([]Validator[T], 0, len(validators)),
	}

	for i := range validators {
		switch v := validators[i].(type) {
		case multiValidator[T]:
			mv.validators = append(mv.validators, v.validators...)
		default:
			mv.validators = append(mv.validators, v)
		}
	}

	return mv
}

// NoOp returns a no-op Validator.
func NoOp[T any]() Validator[T] {
	return noOpValidator[T]{}
}

type noOpValidator[T any] struct{}

// Validate implements the Validator interface.
//
// This is a no-op call and the returned error is always nil.
func (noOpValidator[T]) Validate(T) error { return nil }

// nonNilFunc is a solution for not being able to create a type constraint for
// Func-like functions *and* Validator interface types like so:
//
//	type nullable[T any] interface {
//	  ~func(T) error | Validator[T]
//	}
func nonNilFunc[V any, T ~func(V) error](validators []T) []T {
	squash := make([]T, 0, len(validators))

	for i := range validators {
		if validators[i] == nil {
			continue
		}

		squash = append(squash, validators[i])
	}

	return squash
}

// nonNilValidators is a solution for not being able to create a type constraint for
// Func-like functions *and* Validator interface types like so:
//
//	type nullable[T any] interface {
//	  ~func(T) error | Validator[T]
//	}
func nonNilValidators[T any](validators []Validator[T]) []Validator[T] {
	squash := make([]Validator[T], 0, len(validators))

	for i := range validators {
		if validators[i] == nil {
			continue
		}

		squash = append(squash, validators[i])
	}

	return squash
}
