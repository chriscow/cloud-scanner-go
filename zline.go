package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"
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

// ZeroTypes is a convenience for enumerating all ZeroTypes
var ZeroTypes = []ZeroType{Primes, SixNFives, SixN, Zeta, ZetaNorm1, ZetaNorm2, Comp1, Comp2}

// Zeros are the unique number sequence of type ZeroType. The values may be
// scaled by the scalar value. If Negatives is true, the values are also negated
type Zeros struct {
	ZeroType  ZeroType
	Scalar    float64
	Negatives bool
	Count     int
	Values    []float64 `json:"-"`
}

// ZLine is a number line in space starting at the specified origin and rotated
// about the origin by the specified angle in degrees
type ZLine struct {
	Origin Vector2
	Angle  float64
	Limit  float64
	Zeros  []*Zeros
}

// String returns the string representation of the ZeroType enum
func (z ZeroType) String() string {
	return [...]string{
		"Primes", "SixNFives", "SixN", "Zeta", "ZetaNorm1",
		"ZetaNorm2", "Comp1", "Comp2",
	}[z]
}

// GetZType returns a ZeroType from a string representation
func (z ZeroType) GetZType(name string) (ZeroType, error) {
	switch strings.ToLower(name) {
	case "primes", "prime":
		return Primes, nil
	case "sixnfives", "sixn5s", "6n5s", "sixnfive", "sixn5", "6n5":
		return SixNFives, nil
	case "sixn", "6n":
		return SixN, nil
	case "zeta", "zetas":
		return Zeta, nil
	case "zetanorm1":
		return ZetaNorm1, nil
	case "zetanorm2":
		return ZetaNorm2, nil
	case "comp1", "comp1s":
		return Comp1, nil
	case "comp2", "comp2s":
		return Comp2, nil
	default:
		return 0, errors.New("Unknown zero type")
	}
}

// NewZLine creates and initializes a zline
func NewZLine(origin Vector2, zeros []ZeroType, limit, scale float64, neg bool) (*ZLine, error) {
	zline := &ZLine{
		Origin: origin,
		Limit:  limit,
		Zeros:  make([]*Zeros, 0),
	}

	for _, ztype := range zeros {
		zero, err := LoadZeros(ztype, limit, scale, neg)
		if err != nil {
			return nil, err
		}

		zline.Zeros = append(zline.Zeros, zero)
	}

	return zline, nil
}

// MaxZeroVal returns the largest zero value on the ZLine
func (z *ZLine) MaxZeroVal() float64 {
	var maxZero float64
	for _, zero := range z.Zeros {
		if zero.Values[len(zero.Values)-1] > maxZero {
			maxZero = zero.Values[len(zero.Values)-1]
		}
	}
	return maxZero
}

// LoadZeros loads the numeric values from a data file and returns the indicated
// numeric type up to the maxValue, scaled by the scale value.
// The maxValue is the maximum value loaded before scaling.
func LoadZeros(ztype ZeroType, maxValue, scale float64, neg bool) (*Zeros, error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	path := cwd + "/data/zeros/" + ztype.String() + ".x1.0000"
	values, err := loadLocal(path, maxValue, scale, neg)
	if err != nil {
		return nil, err
	}

	return &Zeros{
		ZeroType:  ztype,
		Scalar:    scale,
		Negatives: neg,
		Count:     len(values),
		Values:    values,
	}, nil
}

func loadLocal(path string, maxValue float64, scale float64, neg bool) ([]float64, error) {
	zeros := make([]float64, 0, 256)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data := make([]float64, 0, 256)
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	for _, value := range data {

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

func max(vals []float64) float64 {
	res := -math.MaxFloat64
	for _, val := range vals {
		if val > res {
			res = val
		}
	}

	return res
}
