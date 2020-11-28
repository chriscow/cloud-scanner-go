package scanner

import "math"

// Vector2 is a 2D Vector2 struct
type Vector2 struct {
	X float64 `msgpack:"x"`
	Y float64 `msgpack:"y"`
}

func Min(a, b Vector2) Vector2 {
	return Vector2{X: math.Min(a.X, b.X), Y: math.Min(a.Y, b.Y)}
}

func Max(a, b Vector2) Vector2 {
	return Vector2{X: math.Max(a.X, b.X), Y: math.Max(a.Y, b.Y)}
}

func (p Vector2) Normalize() Vector2 {
	return p.Div(p.Length())
}

func (p Vector2) Length() float64 {
	return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

func (p Vector2) Div(divisor float64) Vector2 {
	return Vector2{X: p.X / divisor, Y: p.Y / divisor}
}

func (p Vector2) Scale(mul float64) Vector2 {
	return Vector2{X: p.X * mul, Y: p.Y * mul}
}

func (p Vector2) Add(b Vector2) Vector2 {
	return Vector2{X: p.X + b.X, Y: p.Y + b.Y}
}

func (p Vector2) Sub(b Vector2) Vector2 {
	return Vector2{X: p.X - b.X, Y: p.Y - b.Y}
}

func (p Vector2) Dot(b Vector2) float64 {
	return p.X*b.X + p.Y*b.Y
}

func (p Vector2) Distance(pt Vector2) float64 {
	return math.Sqrt((pt.X-p.X)*(pt.X-p.X) + (pt.Y-p.Y)*(pt.Y-p.Y))
}

func (p Vector2) Abs() Vector2 {
	return Vector2{X: math.Abs(p.X), Y: math.Abs(p.Y)}
}
