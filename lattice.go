package main

import (
	"errors"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"

	"github.com/shamaton/msgpack"
)

// Lattice is a set of Points in space arranged in interesting ways.
type Lattice struct {
	LatticeType LatticeType
	VertexType  VertexType
	Parameters  interface{}

	Points []Vector2 `json:"ignore"`
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

// String returns the stringified version of a LatticeType
func (lt LatticeType) String() string {
	return [...]string{
		"Pinwheel", "Fibonacci", "Grid", "Penrose",
	}[lt]
}

// GetLType returns the lattice type from its string representation
func (lt LatticeType) GetLType(name string) (LatticeType, error) {
	switch strings.ToLower(name) {
	case "pinwheel":
		return Pinwheel, nil
	case "fibonacci":
		return Fibonacci, nil
	case "grid":
		return Grid, nil
	case "penrose":
		return Penrose, nil
	default:
		return 0, errors.New("Unknown lattice type")
	}
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

// NewLattice loads or generates lattice Points. If the lattice is generated,
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

// Filter filters out Points that are not candidates for scanning
func (l *Lattice) Filter(origin Vector2, radius float64, maxZero float64, distanceLimit float64) []Vector2 {

	Points := make([]Vector2, 0, len(l.Points))

	// Calculate the radius based on ZLines selected and DistanceLimit
	// Double it because the origin can be at the edge of this
	r := math.Sqrt((radius+maxZero)*(radius+maxZero) + distanceLimit*distanceLimit)

	for _, pt := range l.Points {
		// if its in the radius, copy the Vector2 and move the index
		distance := math.Sqrt((pt.X-origin.X)*(pt.X-origin.X) + (pt.Y-origin.Y)*(pt.Y-origin.Y))
		if math.Abs(distance) <= r {
			Points = append(Points, pt)
		}
	}

	return Points
}

func (l *Lattice) Bounds() BoundingBox {
	return NewBounds(l.Points)
}

// Partition finds all origins that with the given radius, will cover
// the entire lattice
func (l *Lattice) Partition(radius float64) []Vector2 {

	diameter := radius * 2
	hull := makeHull(l.Points)
	bounds := l.Bounds()
	bmax := bounds.Max()
	bmin := bounds.Min()
	origins := make([]Vector2, 0)

	// We are doing the hexogonal tiling with circles over the lattice:
	// https://stackoverflow.com/questions/7716460/fully-cover-a-rectangle-with-minimum-amount-of-fixed-radius-circles

	row := 1
	point := bounds.Min()

	for point.Y <= bmax.Y+radius {
		intersects := circleIntersectsPolygon(hull, point, radius)
		if intersects {
			origins = append(origins, point)
		}

		point = Vector2{X: point.X + diameter, Y: point.Y}

		// If we have reached the right side of the lattice,
		// do a carriage return up
		if point.X-radius > bmax.X {
			if row%2 == 0 {
				point = Vector2{X: bmin.X, Y: point.Y + radius}
			} else {
				point = Vector2{X: bmin.X + radius, Y: point.Y + radius}
			}

			row++
		}
	}

	return origins

}
