package snapshot

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeatureEnabledF(t *testing.T) {

	// defaultValue is the percentage used when the key is missing from the
	// snapshot or the value is invalid.
	//
	// an arbitrary value within the range is used to ensure that the function
	// does not fall back to 0 or 100.
	const defaultValue SampleRate = 69

	type TestCase struct {
		RuntimeConfigPercentage string
		Delta                   float64
		ExpectedPercentage      SampleRate
	}

	var cases []TestCase

	{
		boundaryCases := []string{
			"0",
			"100",
		}
		for _, c := range boundaryCases {
			f, err := strconv.ParseFloat(c, 64)
			if err != nil {
				t.Fatal("malformed test case")
			}
			cases = append(cases, TestCase{
				RuntimeConfigPercentage: c,
				ExpectedPercentage:      SampleRate(f), // should be parsed value
				Delta:                   0,             // percentage should match the expected exactly
			})
		}

		validVals := []string{ // in (0, 100)
			"1",
			"2.5",
			"5",
			"10",
			"20.1",
			"40",
			"80",
			"99",
		}
		for _, c := range validVals {
			f, err := strconv.ParseFloat(c, 64)
			if err != nil {
				t.Fatal("malformed test case")
			}
			cases = append(cases, TestCase{
				RuntimeConfigPercentage: c,
				ExpectedPercentage:      SampleRate(f),
				Delta:                   1, // some deviation is expected
			})
		}

		invalid := []string{
			"-1",
			"101",
			"foo",
		}
		for _, c := range invalid {
			cases = append(cases, TestCase{
				RuntimeConfigPercentage: c,
				ExpectedPercentage:      defaultValue, // should fall back to default
				Delta:                   1,            // some deviation is expected
			})
		}

		misc := []TestCase{
			{
				RuntimeConfigPercentage: "\t50.5    \n", // a valid value with white space
				ExpectedPercentage:      50.5,
				Delta:                   1, // some deviation is expected
			},
		}
		for _, tc := range misc {
			cases = append(cases, tc)
		}
	}

	for _, c := range cases {

		const (
			seed = 1 // for determinism
			key  = "doesntmatter"
		)

		s := NewMock()
		s.Set(key, c.RuntimeConfigPercentage)
		r := rand.New(rand.NewSource(seed))
		percentActual := percentTrue(10000, func() bool {
			return FeatureEnabledF(s, r, key, defaultValue)
		})

		assert.InDelta(t, float64(c.ExpectedPercentage), percentActual, c.Delta, fmt.Sprintln(c))
	}
}

// percentTrue runs |f| n times, returning the [0, 100] rate at which f returns
// true
func percentTrue(n int, f func() bool) float64 {
	var numTrue float64 = 0
	for i := 0; i < n; i++ {
		if f() {
			numTrue += 1
		}
	}
	return numTrue / float64(n) * 100
}
