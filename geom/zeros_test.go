package geom

import (
	"os"
	"testing"
)

func TestLoadLoad(t *testing.T) {
	const maxval = 100
	const scalar = .3

	for _, zt := range ZeroTypes {
		zeros := Zeros{
			ZeroType:  zt,
			Scalar:    scalar,
			Negatives: false,
		}

		err := LoadZeros(&zeros, maxval)
		if err != nil {
			t.Log("failed to load zeros for", zt.String(), err)
			t.Fail()
		}

		if len(zeros.Values) == 0 {
			t.Log("no zeros returned for type", zt.String())
			t.Fail()
		}

		if zeros.Values[len(zeros.Values)-1] > maxval/scalar {
			t.Log("zero value greater than maximum specified")
			t.Fail()
		}
	}
}

func TestAllTypesHaveStrings(t *testing.T) {
	for _, zt := range ZeroTypes {
		s := zt.String() // will blow up with index out of range

		result := false
		switch zt {
		case Primes:
			result = s == "Primes"
		case SixNFives:
			result = s == "SixNFives"
		case SixN:
			result = s == "SixN"
		case Zeta:
			result = s == "Zeta"
		case ZetaNorm1:
			result = s == "ZetaNorm1"
		case ZetaNorm2:
			result = s == "ZetaNorm2"
		case Comp1:
			result = s == "Comp1"
		case Comp2:
			result = s == "Comp2"
		}

		if result == false {
			t.Log("unexpected string / type for: ", s)
			t.Fail()
		}
	}
}
