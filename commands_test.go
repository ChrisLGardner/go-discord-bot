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
			"chris : 2021-01-02 12:00:00 +0000 GMT",
			false,
		},
		{
			"sarah",
			"sarah : 2021-01-02 07:00:00 -0500 EST",
			false,
		},
		{
			"dave",
			"dave : 2021-01-02 20:00:00 +0800 AWST",
			false,
		},
		{
			"mary rose",
			"mary rose : 2021-01-02 04:00:00 -0800 PST",
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
