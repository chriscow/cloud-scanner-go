package geom

import "testing"

func TestBoundsNilReturnsDefault(t *testing.T) {
	b := NewBounds(nil)

	if b.Center.X != 0 || b.Center.Y != 0 || b.Extents.X != 0 || b.Extents.Y != 0 {
		t.Log("expected default (zeroed) bounds. got:", b.Center, b.Extents)
		t.Fail()
	}
}
