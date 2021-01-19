package main

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func init() {
	start := time.Now()
	os.Setenv("APP_DATA", "/Users/chris/pasta/go-scanner/data/")
	if err := loadOEIS(); err != nil {
		log.Fatal("loadOEIS: ", err)
	}

	if err := loadOEISDecExp(); err != nil {
		log.Fatal("loadOEISDecExp: ", err)
	}

	fmt.Println("loaded oeis data in", time.Since(start))
}

func TestSeqExists(t *testing.T) {
	seq := oeisSeq["A000012"]
	if seq == nil {
		t.Fatal("Sequence not found: `A000012`")
	}

	check := []int{1, 1, 1, 1, 1, 1, 1, 1}

	if !equal(seq[:8], check) {
		t.Fatal("Unexpected sequence: ", seq[:8])
	}
}

func TestDecimalExpansionDescExists(t *testing.T) {
	desc := decExpNames["A000796"] // decimal expansion of Pi
	if desc == "" {
		t.Fatal("Sequence description not found: `A000796`")
	}

	if desc != "Decimal expansion of Pi (or digits of Pi)." {
		t.Fatal("Unexpected description for `A000796`: ", desc)
	}
}

func TestNoSequenceExists(t *testing.T) {
	// just some random numbers. No sequence exists for this as of this writing
	digits := []int{1, 9, 2, 7, 8, 3, 4, 5, 3}
	pos := 0
	deOnly := false

	data, err := searchOEIS(digits, pos, deOnly)
	if err != nil {
		t.Fatal("Unexpected error while searching OEIS:", err)
	}

	if len(data) != 0 {
		t.Fatal("Expected no results but found:", len(data))
	}
}

// TestSequenceAtStart finds matching sequences that start with the digits provided.
func TestSequenceAtStart(t *testing.T) {
	// just some random numbers. No sequence exists for this as of this writing
	digits := []int{2, 3, 4, 5, 7}
	pos := 0
	deOnly := false

	data, err := searchOEIS(digits, pos, deOnly)
	if err != nil {
		t.Fatal("Unexpected error while searching OEIS:", err)
	}

	if len(data) == 0 {
		t.Fatal("Expected results but found none")
	}

	for i := range data {
		seq := data[i]

		if len(seq.Seq) < len(digits) {
			t.Fatal("Expected sequence to be same or longer than digits query", len(seq.Seq), len(digits))
		}

		// compare the first len(digits) of the sequence returned
		// and make sure they are the same.
		if !equal(digits, seq.Seq[:len(digits)]) {
			t.Fatal("Expected to find digits at beginning of sequence", digits, "!=", seq.Seq[:len(digits)])
		}
	}
}

func TestSequenceInMiddleMultiResult(t *testing.T) {
	digits := []int{10, 11, 12, 13} // A000037
	pos := 6
	deOnly := false

	data, err := searchOEIS(digits, pos, deOnly)
	if err != nil {
		t.Fatal("Unexpected error while searching OEIS:", err)
	}

	if len(data) < 100 {
		t.Error("Expected more than 100 results but", len(data), "found")
	}

	for i := range data {
		seq := data[i]

		if len(seq.Seq) < len(digits) {
			t.Fatal("Expected sequence to be same or longer than digits query", len(seq.Seq), len(digits))
		}

		// compare the first len(digits) of the sequence returned
		// and make sure they are the same.
		s := seq.Seq[pos : pos+len(digits)]
		if len(s) != len(digits) {
			t.Fatal("Expected slice of sequence to be same length as digits")
		}

		if !equal(digits, s) {
			t.Fatal("Expected to find digits at beginning of sequence", digits, "!=", s)
		}
	}
}
func TestSequenceInMiddleSingleResult(t *testing.T) {
	// just some random numbers. Only one sequence exists as of this writing
	digits := []int{-33, 36, -46, 51, -53, 58, -68} // A000025
	pos := 20
	deOnly := false

	data, err := searchOEIS(digits, pos, deOnly)
	if err != nil {
		t.Log("Unexpected error while searching OEIS:", err)
		t.FailNow()
	}

	if len(data) != 1 {
		t.Error("Expected one result but", len(data), "found")
	}

	seq := data[0]

	if len(seq.Seq) < len(digits) {
		t.Fatal("Expected sequence to be same or longer than digits query", len(seq.Seq), len(digits))
	}

	// compare the first len(digits) of the sequence returned
	// and make sure they are the same.
	s := seq.Seq[pos : pos+len(digits)]
	if len(s) != len(digits) {
		t.Fatal("Expected slice of sequence to be same length as digits")
	}

	if !equal(digits, s) {
		t.Fatal("Expected to find digits at beginning of sequence", digits, "!=", s)
	}
}

func TestDecimalExpansion(t *testing.T) {

}

// this test times out
func TestDecimalExpansionSingleDigitsOnly(t *testing.T) {
	count := 0

	for k, v := range decExpNames {
		t.Log(k, v)

		count++

		if count > 100 {
			t.Fail() // bail out to see logs
		}

		// Search for the exact sequence as a decimal expansion
		// data, err := searchOEIS(oeisSeq[k][:10], 0, true)
		// if err != nil {
		// 	t.Fatal("Unexpected error while searching OEIS:", err)
		// }

		// // some decimal expension descriptions don't meet all criteria so
		// // we will get zero
		// if len(data) == 0 {
		// 	continue
		// }

		// if len(data) > 1 {
		// 	t.Error("Expected one result but", len(data), "found")
		// }

		// seq := data[0]

		// // check that all the digits are single-digit
		// for i := range seq.Seq {
		// 	if seq.Seq[i] >= 10 {
		// 		t.Log(seq.Seq)
		// 		t.Fatal(seq.ID, "decimal expansion returned multi-digit at index", i, " : ", seq.Seq[i])
		// 	}
		// }
	}
}

func TestMultiDigitRegression(t *testing.T) {
	digits := []int{1, 1, 3, 4, 5, 6, 7, 7, 8, 8, 9, 9, 9, 9} // should not find this
	pos := 0
	deOnly := true

	data, err := searchOEIS(digits, pos, deOnly)
	if err != nil {
		t.Fatal("Unexpected error while searching OEIS:", err)
	}

	if len(data) != 0 {
		t.Fatal("Expected no results but found", len(data))
	}

	for i := range data {
		seq := data[i]
		t.Log(seq)
		// check that all the digits are single-digit
		for i := range seq.Seq {
			if seq.Seq[i] >= 10 {
				t.Log(seq.Seq)
				t.Fatal(seq.ID, "decimal expansion returned multi-digit at index", i, " : ", seq.Seq[i])
			}
		}
	}

	digits = []int{4, 5, 6, 4, 3, 4, 8, 1, 9, 1, 4, 6, 7, 8, 3, 6, 2, 3, 8, 4, 8, 1, 4, 0, 5, 8, 4, 4} // should find this
	data, err = searchOEIS(digits, pos, deOnly)
	if err != nil {
		t.Fatal("Unexpected error while searching OEIS:", err)
	}

	if len(data) != 1 {
		t.Fatal("Expected 1 result but found", len(data))
	}

	for i := range data {
		seq := data[i]
		t.Log(seq)
		// check that all the digits are single-digit
		for i := range seq.Seq {
			if seq.Seq[i] >= 10 {
				t.Log(seq.Seq)
				t.Fatal(seq.ID, "decimal expansion returned multi-digit at index", i, " : ", seq.Seq[i])
			}
		}
	}

}

// Equal tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
