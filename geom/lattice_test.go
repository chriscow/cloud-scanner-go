package geom

import (
	"os"
	"testing"
)

func TestCanLoadPinwheelVertices(t *testing.T) {

	lattice, err := loadLattice(Pinwheel, Vertices)
	if err != nil {
		t.Fatal(err)
	}

	if lattice.LatticeType != Pinwheel {
		t.Log("expected lattice type == Pinwheel but was", lattice.LatticeType.String())
		t.Fail()
	}

	if lattice.VertexType != Vertices {
		t.Log("expexted vertex type == Vertices but was", lattice.VertexType.String())
		t.Fail()
	}
	if len(lattice.Points) != 277845 {
		t.Log("expected pinwheel verticies to have 277845 Points but it had", len(lattice.Points))
		t.Fail()
	}
}
