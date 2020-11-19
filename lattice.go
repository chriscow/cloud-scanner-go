package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"strings"
	"sync"

	"github.com/shamaton/msgpack"
)

// Lattice points
type Lattice struct {
	LatticeType LatticeType
	VertexType  VertexType
	Points      []Point `json:"Points"`
}

type fileLattice struct {
	Points []float64 `json:"Points"`
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

// LoadLattice loads and returns a *Lattice
func LoadLattice(ltype LatticeType, vtype VertexType) (*Lattice, error) {
	lstr := strings.ToLower(ltype.String())
	vstr := strings.ToLower(vtype.String())
	path := "./data/lattices/" + lstr + "." + vstr + ".msgpack"
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

// LoadJSONFile loads lattice points from the legacy JSON file format
func loadLegacyLattice(path string) ([]Point, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	s := new(sync.WaitGroup)
	m := new(sync.Mutex)
	points := make([]Point, 0, 1024*1024)

	s.Add(len(files))
	for _, f := range files {
		go func(name string) {
			const pointCount = 8192

			data, err := ioutil.ReadFile(path + name)
			if err != nil {
				log.Fatal(err)
			}

			var lp fileLattice
			if err := json.Unmarshal(data, &lp); err != nil {
				log.Fatal(err)
			}

			// Load all the points into the pts buffer then append that
			// buffer into our main lattice point array separately while locked
			pts := make([]Point, 0, pointCount)
			for i := 0; i < len(lp.Points); i += 2 {
				pt := Point{
					X: lp.Points[i],
					Y: lp.Points[i+1],
				}

				pts = append(pts, pt)
			}

			m.Lock()
			points = append(points, pts...)
			m.Unlock()

			s.Done()
		}(f.Name())
	}

	s.Wait()
	return points, nil
}

// Filter filters out points that are not candidates for scanning
func (l *Lattice) Filter(origin Point, radius float64, zeros []float64, distanceLimit float64) {

	maxZero := max(zeros)
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

	l.Points = points
	return
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
