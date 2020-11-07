package gobinance

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestMillisTimestamp_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		jsonInput      string
		errorExpected  bool
		expectedResult time.Time
	}{
		{
			jsonInput:      "1234567890123",
			errorExpected:  false,
			expectedResult: time.Date(2009, 02, 13, 23, 31, 30, int(123*time.Millisecond), time.UTC),
		},
		{
			jsonInput:  "0",
			errorExpected:  false,
			expectedResult: time.Date(1970,01,01,00,00,00,00,time.UTC),
		},
		{
			jsonInput:  "-1234567890123",
			errorExpected: false,
			expectedResult: time.Date(1930,11,18,00,28,29,int(time.Second - (123 * time.Millisecond)), time.UTC),
		},
		{
			jsonInput:  `"invalid"`,
			errorExpected: true,
		},
	}

	for _, tc :=range testCases{
		tc := tc
		t.Run(fmt.Sprintf("unmarshalling %s",tc.jsonInput), func(t *testing.T) {
			var uut millisTimestamp
			err := json.Unmarshal( []byte(tc.jsonInput), &uut)
			if tc.errorExpected {
				if err == nil {
					t.Errorf("error expected but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got := time.Time(uut); !reflect.DeepEqual(got, tc.expectedResult) {
				t.Errorf("unexpected time. expected\n\t%s (%#v)\ngot\n\t%s (%#v)",
					tc.expectedResult.Format(time.RFC3339Nano),
					tc.expectedResult,
					got.Format(time.RFC3339Nano),
					got,
				)
			}
		})
	}
}
