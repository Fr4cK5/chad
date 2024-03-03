package parse_test

import (
	"fmt"
	"testing"

	"github.com/Fr4cK5/chad/internal/parse"
)

func TestParse(t *testing.T) {
	
	test_cases := []string{
		`You really should listen to Innerbloom by Rüfüs Du Sol`,
		`'that is crazy' -ab 'Assigned to b alone' --use-postfix -go yo "Hey there! >.<"`,
		`--single-arg hello -stack world`,
	}

	for i, test := range test_cases {
		result := parse.ParseFromString(test)
		fmt.Println("Test case", i, result)
	}
}

