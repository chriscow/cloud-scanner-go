package geom

import "math"

type BoundingBox struct {
	// Center of the bounding box
	Center Vector2

	// Extents is like the radius. Half the total size
	Extents Vector2
}

func NewBounds(points []Vector2) BoundingBox {

	var min, max Vector2

	for _, pt := range points {
		min.X = math.Min(pt.X, min.X)
		min.Y = math.Min(pt.Y, min.Y)
		max.X = math.Max(pt.X, max.X)
		max.Y = math.Max(pt.Y, max.Y)
	}

	return BoundingBox{
		Center:  max.Add(min).Div(2),
		Extents: max.Sub(min).Abs().Div(.5),
	}
}

func (b BoundingBox) Min() Vector2 {
	return b.Center.Sub(b.Extents)
}

func (b BoundingBox) Max() Vector2 {
	return b.Center.Add(b.Extents)
}

func (b BoundingBox) Size() Vector2 {
	return b.Extents.Scale(2)
}
