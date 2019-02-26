package snapshot

import (
	"math"
	"strconv"
	"strings"

	"errors"
)

type SampleRate float64

const (
	Off SampleRate = 0
	Max SampleRate = 100
)

type Rand interface {
	// Float64 has the same contract as rand.Rand.Float64 in the standard
	// library
	Float64() float64
}

func (s SampleRate) MultipliedBy(r SampleRate) SampleRate {
	p := float64(s) * float64(r) / 100
	return SampleRate(p)
}

func (s SampleRate) String() string {
	return strconv.FormatFloat(float64(s), 'f', 6, 64)
}

func GetSampleRateOrDefault(runtime IFace, key string, defaultValue SampleRate) SampleRate {
	s, err := GetSampleRate(runtime, key)
	if err != nil {
		return defaultValue
	}

	return s
}

func IsSampleRateDefined(runtime IFace, key string) bool {
	_, err := GetSampleRate(runtime, key)
	return err == nil
}

func GetSampleRate(runtime IFace, key string) (SampleRate, error) {
	if runtime.Get(key) == "" {
		return Off, errors.New("Key does not exist")
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(runtime.Get(key)), 64)
	if err != nil {
		return Off, err
	}

	if parsed < 0 || parsed > 100 {
		return Off, errors.New("Invalid sample rate")
	}

	return SampleRate(parsed), nil
}

// FeatureEnabledF extracts a float |value| from a runtime snapshot given a
// string key and returns true |value| percent of the time.
//
// NB: supports floating point granularity.
//
// NB: if a value cannot be found for the given runtime |key|, this function
// returns true |defaultValue| percent of the time
func FeatureEnabledF(runtime IFace, r Rand, key string, defaultValue SampleRate) bool {
	parsed := GetSampleRateOrDefault(runtime, key, defaultValue)
	return FeatureEnabled(r, parsed)
}

func FeatureEnabled(r Rand, s SampleRate) bool {
	return r.Float64()*100 < math.Min(float64(s), 100)
}
