package iam

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viebiz/lit/mocks/mockcasbin"
)

func TestEnforcer_Enforce(t *testing.T) {
	type args struct {
		sub, obj, act string
	}

	tcs := map[string]struct {
		args      args
		mockAllow bool
		mockErr   error
		wantErr   error
	}{
		"allowed": {
			args:      args{sub: "alice", obj: "data1", act: "read"},
			mockAllow: true,
			mockErr:   nil,
			wantErr:   nil,
		},
		"denied": {
			args:      args{sub: "bob", obj: "data2", act: "write"},
			mockAllow: false,
			mockErr:   nil,
			wantErr:   ErrActionIsNotAllowed,
		},
		"cb error": {
			args:      args{sub: "eve", obj: "data3", act: "delete"},
			mockAllow: false,
			mockErr:   errors.New("cb-failed"),
			wantErr:   errors.New("cb-failed"),
		},
	}

	for scenario, tc := range tcs {
		t.Run(scenario, func(t *testing.T) {
			// Given
			mockEnf := new(mockcasbin.MockIEnforcer)
			mockEnf.On("Enforce", tc.args.sub, tc.args.obj, tc.args.act).
				Return(tc.mockAllow, tc.mockErr)

			// When
			e := enforcer{cb: mockEnf}
			err := e.Enforce(tc.args.sub, tc.args.obj, tc.args.act)

			// Then
			if tc.wantErr != nil {
				require.EqualError(t, err, tc.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
			mockEnf.AssertExpectations(t)
		})
	}
}
