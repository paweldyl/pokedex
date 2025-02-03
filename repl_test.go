package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct{
		input    string
		expected []string
	}{
		{
			input: "  hello world   ",
			expected: []string{"hello", "world"},
		},
		{
			input: "  hello my world   ",
			expected: []string{"hello", "my", "world"},
		},	
	}

	for _, testCase := range cases{
		actualResult := CleanInput(testCase.input)
		if len(actualResult) != len(testCase.expected) {
			t.Errorf("incorrect array size")
		}
		if len(actualResult) != len(testCase.expected) {
			t.Errorf("incorrect array size")
		}
		for i := 0; i < len(actualResult); i++ {
			expectedArrayItem := testCase.expected[i]
			returnedArrayItem := actualResult[i]

			if expectedArrayItem != returnedArrayItem {
				t.Errorf("different items, %v != %v",expectedArrayItem, returnedArrayItem)
			}
		}
	}
}