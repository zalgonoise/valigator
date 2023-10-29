package valigator

import (
	"context"
	"testing"
)

func TestNewContext(t *testing.T) {
	var (
		testNoOp = func(context.Context, testConfig) error { return nil }
	)

	for _, testcase := range []struct {
		name     string
		input    []func(context.Context, testConfig) error
		isNoOp   bool
		wantsLen int
	}{
		{
			name:   "None",
			isNoOp: true,
		},
		{
			name: "One",
			input: []func(ctx context.Context, config testConfig) error{
				testNoOp,
			},
		},
		{
			name: "Three",
			input: []func(ctx context.Context, config testConfig) error{
				testNoOp, testNoOp, testNoOp,
			},
			wantsLen: 3,
		},
		{
			name: "WithNilsOneIsValid",
			input: []func(ctx context.Context, config testConfig) error{
				nil, nil, testNoOp, nil, nil,
			},
		},
		{
			name: "AllNil",
			input: []func(ctx context.Context, config testConfig) error{
				nil, nil, nil, nil, nil,
			},
			isNoOp: true,
		},
		{
			name: "ThreeWithNils",
			input: []func(ctx context.Context, config testConfig) error{
				testNoOp, nil, testNoOp, nil, testNoOp, nil,
			},
			wantsLen: 3,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			validator := NewContext(testcase.input...)

			_ = validator.Validate(context.Background(), testConfig{})

			switch v := validator.(type) {
			case noOpContextValidator[testConfig]:
				requireEqual(t, true, testcase.isNoOp)
				return
			case multiContextValidator[testConfig]:
				requireEqual(t, testcase.wantsLen, len(v.validators))
				return
			}

			requireEqual(t, false, testcase.isNoOp)
		})
	}
}

func TestJoinContext(t *testing.T) {
	var testNoOp = ContextFunc[testConfig](func(context.Context, testConfig) error { return nil })

	for _, testcase := range []struct {
		name     string
		input    []ContextValidator[testConfig]
		isNoOp   bool
		wantsLen int
	}{
		{
			name:   "None",
			isNoOp: true,
		},
		{
			name: "One",
			input: []ContextValidator[testConfig]{
				testNoOp,
			},
		},
		{
			name: "Three",
			input: []ContextValidator[testConfig]{
				testNoOp, testNoOp, testNoOp,
			},
			wantsLen: 3,
		},
		{
			name: "WithNilsOneIsValid",
			input: []ContextValidator[testConfig]{
				nil, nil, testNoOp, nil, nil,
			},
		},
		{
			name: "AllNil",
			input: []ContextValidator[testConfig]{
				nil, nil, nil, nil, nil,
			},
			isNoOp: true,
		},
		{
			name: "ThreeWithNils",
			input: []ContextValidator[testConfig]{
				testNoOp, nil, testNoOp, nil, testNoOp, nil,
			},
			wantsLen: 3,
		},
		{
			name: "ThreeWithAMultiValidator",
			input: []ContextValidator[testConfig]{
				JoinContext[testConfig](testNoOp, testNoOp, testNoOp),
				nil, testNoOp, nil, testNoOp, nil,
			},
			wantsLen: 5,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			validator := JoinContext(testcase.input...)

			_ = validator.Validate(context.Background(), testConfig{})

			switch v := validator.(type) {
			case noOpContextValidator[testConfig]:
				requireEqual(t, true, testcase.isNoOp)
				return
			case multiContextValidator[testConfig]:
				requireEqual(t, testcase.wantsLen, len(v.validators))
				return
			}

			requireEqual(t, false, testcase.isNoOp)
		})
	}
}
