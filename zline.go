package main

// ZLine can contain multiple zeros
type ZLine struct {
	Origin Point
	Angle  float64
	Length int
	Zeros  []Zeros
}
