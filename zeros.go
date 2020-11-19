package main

import (
	"encoding/json"
	"io/ioutil"
)

// ZeroType enumeration defines what type of numbers the values represent
type ZeroType int

const (
	// Primes are prime numbers
	Primes ZeroType = iota
	// SixNFives are all numbers 6n - 1, not divisible by 5 and also not prime
	SixNFives
	// SixN are all numbers 6n - 1 except prime numbers
	SixN
	// Zeta zeros of the Reimann-Zeta function
	Zeta
	// ZetaNorm1 are zeta * Math.Log(zeta) / (2d * Math.PI)
	ZetaNorm1
	// ZetaNorm2 are even more complicated
	ZetaNorm2
	// Comp1 are all whole numbers starting at 4 that aren't primes
	Comp1
	// Comp2 are all mi,bers starting at 4 that aren't in SixN or Prime
	Comp2
)

// Zeros are the unique number sequence of type ZeroType. The values may be
// scaled by the scalar value. If Negatives is true, the values are also negated
type Zeros struct {
	ZeroType  ZeroType
	Scalar    float64
	Negatives bool
	Values    []float64
}

// String returns the string representation of the ZeroType enum
func (z ZeroType) String() string {
	return [...]string{
		"Primes", "SixNFives", "SixN", "Zeta", "ZetaNorm1",
		"ZetaNorm2", "Comp1", "Comp2",
	}[z]
}

// LoadZeros returns the indicated zero type up to the maxValue, scaled by the
// scale value.  The maxValue is before scaling.
func LoadZeros(ztype ZeroType, maxValue float64, scale float64, neg bool) (*Zeros, error) {
	path := "./data/zeros/" + ztype.String() + ".x1.0000"
	values, err := loadLocal(path, maxValue, scale, neg)
	if err != nil {
		return nil, err
	}

	return &Zeros{
		ZeroType:  ztype,
		Scalar:    scale,
		Negatives: neg,
		Values:    values,
	}, nil
}

// LoadLocal loads zeros from a local JSON file with the values
func loadLocal(path string, maxValue float64, scale float64, neg bool) ([]float64, error) {
	zeros := make([]float64, 0, 256)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &zeros); err != nil {
		return nil, err
	}

	for _, value := range zeros {

		if value > maxValue {
			break
		}

		zeros = append(zeros, value*scale)
		if neg {
			zeros = append(zeros, -value*scale)
		}
	}

	return zeros, nil
}
