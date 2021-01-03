package main

import (
	"context"
	"os"
	"testing"
	"time"
)

type TestTimeItem struct {
	input    string
	result   string
	hasError bool
}

func TestGetTime(t *testing.T) {

	json := "{\"chris\":\"GMT\",\"sarah\":\"EST\",\"dave\":\"Australia/Perth\",\"mary rose\":\"US/Pacific\"}"
	os.Setenv("MEMBER_TIMEZONES", json)

	parsed, _ := time.Parse(time.RFC3339, "2021-01-02T12:00:00Z")

	testCases := []TestTimeItem{
		{
			"",
			"no user specified",
			false,
		},
		{
			"chris",
			"chris : 12:00, 2 January 2021, (GMT)",
			false,
		},
		{
			"sarah",
			"sarah : 07:00, 2 January 2021, (EST)",
			false,
		},
		{
			"dave",
			"dave : 20:00, 2 January 2021, (Australia/Perth)",
			false,
		},
		{
			"mary rose",
			"mary rose : 04:00, 2 January 2021, (US/Pacific)",
			false,
		},
		{
			"no one",
			"User not found",
			true,
		},
	}

	for _, test := range testCases {

		ctx := context.Background()

		res, err := getTime(ctx, parsed, test.input)

		if test.hasError == true {
			if err.Error() != test.result {
				t.Errorf("getTime with args %v: FAILED, expected %v but got %v", test.input, test.result, err)
			}
		} else {
			if res != test.result {
				t.Errorf("getTime with args %v: FAILED, expected %v but got %v", test.input, test.result, res)
			}
		}
	}
}
