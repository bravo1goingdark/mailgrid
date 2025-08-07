package utils

import (
	"reflect"
	"testing"

	utils2 "github.com/bravo1goingdark/mailgrid/utils"
)

func TestSplitAndTrim(t *testing.T) {
	input := "  a@example.com , b@example.com, ,c@example.com, "
	expected := []string{"a@example.com", "b@example.com", "c@example.com"}

	result := utils2.SplitAndTrim(input)
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestSplitAndTrim_Empty(t *testing.T) {
	if got := utils2.SplitAndTrim(""); len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}
