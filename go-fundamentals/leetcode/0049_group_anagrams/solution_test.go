package groupanagrams

import (
	"reflect"
	"testing"
)

func TestGroupAnagrams(t *testing.T) {
	input := []string{"eat", "tea", "tan", "ate", "nat", "bat"}
	expected := [][]string{{"eat", "tea", "ate"}, {"tan", "nat"}, {"bat"}}
	result := GroupAnagrams(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("GroupAnagrams(%v) = %v; Expected: %v", input, result, expected)
	}
}