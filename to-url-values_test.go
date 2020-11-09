package gobinance

import (
	"github.com/google/go-cmp/cmp"
	"math/big"
	"net/url"
	"testing"
)

func TestToURLValues(t *testing.T) {
	testCases := []struct {
		name           string
		input          interface{}
		errorExpected  bool
		expectedOutput url.Values
	}{
		{
			name:          "non-struct or map",
			input:         123,
			errorExpected: true,
		},
		{
			name:           "empty struct",
			input:          struct{}{},
			errorExpected:  false,
			expectedOutput: make(url.Values),
		},
		{
			name: "struct with unexported fields",
			input: struct {
				iAmPrivate string
				IAmPublic  string
			}{
				iAmPrivate: "privateField",
				IAmPublic:  "publicField",
			},
			expectedOutput: url.Values{
				"IAmPublic": []string{"publicField"},
			},
		},
		{
			name: "omitting fields with `-` as their alias",
			input: struct {
				OmitMe string `param:"-"`
				ShowMe string
			}{
				OmitMe: "whoops...",
				ShowMe: "someValue",
			},
			expectedOutput: url.Values{
				"ShowMe": []string{"someValue"},
			},
		},
		{
			name: "aliasing fields",
			input: struct {
				SomeName string `param:"anothername"`
			}{
				SomeName: "someValue",
			},
			expectedOutput: url.Values{
				"anothername": []string{"someValue"},
			},
		},
		{
			name: "including non-string values",
			input: struct {
				BigFloat *big.Float
				Int      int
				Float64  float64
			}{
				BigFloat: big.NewFloat(1.25),
				Int:      1234,
				Float64:  2.5,
			},
			expectedOutput: url.Values{
				"BigFloat": []string{"1.25"},
				"Int":      []string{"1234"},
				"Float64":  []string{"2.5"},
			},
		},
		{
			name: "omitempty",
			input: struct {
				EmptyInt         int        `param:"EmptyInt,omitempty"`
				EmptyBigFloat    *big.Float `param:"EmptyBigFloat,omitempty"`
				EmptyString      string     `param:"EmptyString,omitempty"`
				NonEmptyInt      int        `param:"NonEmptyInt,omitempty"`
				NonEmptyBigFloat *big.Float `param:"NonEmptyBigFloat,omitempty"`
				NonEmptyString   string     `param:"NonEmptyString,omitempty"`
			}{
				NonEmptyInt:      1,
				NonEmptyBigFloat: big.NewFloat(2),
				NonEmptyString:   "three",
			},
			expectedOutput: url.Values{
				"NonEmptyInt":      []string{"1"},
				"NonEmptyBigFloat": []string{"2"},
				"NonEmptyString":   []string{"three"},
			},
		},
		{
			name: "zero-values without omitempty",
			input: struct {
				EmptyInt      int        `param:"EmptyInt"`
				EmptyBigFloat *big.Float `param:"EmptyBigFloat"`
				EmptyString   string     `param:"EmptyString"`
			}{},
			expectedOutput: url.Values{
				"EmptyInt":      []string{"0"},
				"EmptyBigFloat": []string{"<nil>"},
				"EmptyString":   []string{""},
			},
		},
		{
			name: "zero-values with emptyvalue",
			input: struct {
				EmptyBigFloat *big.Float `param:"EmptyBigFloat" emptyvalue:"someString"`
			}{},
			expectedOutput: url.Values{
				"EmptyBigFloat": []string{"someString"},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := toURLValues(tc.input)
			if tc.errorExpected {
				if err == nil {
					t.Errorf("expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.expectedOutput, got); diff != "" {
				t.Errorf("unexpected result:\n%v", diff)
			}
		})
	}
}
