package geom

import (
	"os"
	"testing"
)

func TestNewZLineSingle(t *testing.T) {
	origin := Vector2{}

	for _, ztype := range ZeroTypes {
		z := []ZeroType{ztype}
		zline, err := NewZLine(origin, z, 100, 1, false, 7)
		if err != nil {
			t.Log("NewZLine:", err)
			t.Fail()
		}

		if zline.Limit != 100 {
			t.Log("expected limit to be 100 was", zline.Limit)
			t.Fail()
		}

		if zline.Origin.X != 0 || zline.Origin.Y != 0 {
			t.Log("expected zero vector for origin but got", zline.Origin)
			t.Fail()
		}

		if len(zline.Zeros) != 1 {
			t.Log("expected number of zeros to be 1 but was", len(zline.Zeros))
			t.Fail()
		}

		if zline.Angle != 7 {
			t.Log("expected angle to be 7 but was", zline.Angle)
			t.Fail()
		}
	}
}
