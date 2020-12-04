package geom

// ZLine is a number line in space starting at the specified origin and rotated
// about the origin by the specified angle in degrees
type ZLine struct {
	Origin Vector2
	Angle  float64
	Limit  float64
	Zeros  []Zeros
}

// NewZLine creates and initializes a zline
func NewZLine(origin Vector2, zeros []ZeroType, limit, scale float64, neg bool, angle float64) (ZLine, error) {
	zline := ZLine{
		Origin: origin,
		Limit:  limit,
		Angle:  angle,
		Zeros:  make([]Zeros, 0),
	}

	for _, ztype := range zeros {
		zeros := Zeros{
			ZeroType:  ztype,
			Scalar:    scale,
			Negatives: neg,
		}
		err := LoadZeros(&zeros, limit)
		if err != nil {
			return zline, err
		}

		zline.Zeros = append(zline.Zeros, zeros)
	}

	return zline, nil
}

// MaxZeroVal returns the largest zero value on the ZLine
func (z ZLine) MaxZeroVal() float64 {
	var maxZero float64
	for _, zero := range z.Zeros {
		if zero.Values[len(zero.Values)-1] > maxZero {
			maxZero = zero.Values[len(zero.Values)-1]
		}
	}
	return maxZero
}
