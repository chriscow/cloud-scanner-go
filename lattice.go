package main

import (
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"

	"github.com/shamaton/msgpack"
)

// Lattice is a set of points in space arranged in interesting ways.
type Lattice struct {
	LatticeType LatticeType
	VertexType  VertexType
	Parameters  interface{}

	Points []Point `json:"ignore"`
}

// LatticeType enumeration
type LatticeType int

// VertexType enumeration for vertices of a lattice
type VertexType int

const (
	// Pinwheel lattice
	Pinwheel LatticeType = iota

	// Fibonacci lattice
	Fibonacci

	// Grid lattice
	Grid

	// Penrose lattice
	Penrose
)

func (lt LatticeType) String() string {
	return [...]string{
		"Pinwheel", "Fibonacci", "Grid", "Penrose",
	}[lt]
}

const (
	// Vertices vertex type
	Vertices VertexType = iota

	// Centers vertex type
	Centers
)

func (vt VertexType) String() string {
	return [...]string{
		"Vertices", "Centers",
	}[vt]
}

// NewLattice loads or generates lattice points. If the lattice is generated,
// the default lattice parameters are used
func NewLattice(ltype LatticeType, vtype VertexType) (*Lattice, error) {
	// we will eventually generate lattices but the current set are loaded
	// from files
	return loadLattice(ltype, vtype)
}

func loadLattice(ltype LatticeType, vtype VertexType) (*Lattice, error) {
	lstr := strings.ToLower(ltype.String())
	vstr := strings.ToLower(vtype.String())

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	path := cwd + "/data/lattices/" + lstr + "." + vstr + ".msgpack"
	l := &Lattice{}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := msgpack.Decode(b, l); err != nil {
		return nil, err
	}

	return l, nil
}

// Filter filters out points that are not candidates for scanning
func (l *Lattice) Filter(origin Point, radius float64, maxZero float64, distanceLimit float64) []Point {

	points := make([]Point, 0, len(l.Points))

	// Calculate the radius based on ZLines selected and DistanceLimit
	// Double it because the origin can be at the edge of this
	r := math.Sqrt((radius+maxZero)*(radius+maxZero) + distanceLimit*distanceLimit)

	for _, pt := range l.Points {
		// if its in the radius, copy the point and move the index
		distance := math.Sqrt((pt.X-origin.X)*(pt.X-origin.X) + (pt.Y-origin.Y)*(pt.Y-origin.Y))
		if math.Abs(distance) <= r {
			points = append(points, pt)
		}
	}

	return points
}
