package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithLogger(t *testing.T) {

	type testCase[T interface{}, U interface{}] struct {
		handler     Handler[T, U]
		checkResult func(t *testing.T, output U, err error)
		name        string
	}

	testCases := []testCase[inputEvent, outputEvent]{
		{
			name: "Handler returns result",
			handler: func(ctx context.Context, event inputEvent) (outputEvent, error) {
				return outputEvent{Bar: 1}, nil
			},
			checkResult: func(t *testing.T, output outputEvent, err error) {
				assert.Nil(t, err)
				assert.Equal(t, outputEvent{Bar: 1}, output)
			},
		},
		{
			name: "Handler returns error",
			handler: func(ctx context.Context, event inputEvent) (outputEvent, error) {
				return outputEvent{}, errors.New("something bad happened")
			},
			checkResult: func(t *testing.T, output outputEvent, err error) {
				assert.NotNil(t, err)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			wrappedHandler := WithLogger(tc.handler)
			output, err := wrappedHandler(context.Background(), inputEvent{Foo: 1})
			tc.checkResult(t, output, err)
		})
	}
}

type inputEvent struct {
	Foo int
}

type outputEvent struct {
	Bar int
}
