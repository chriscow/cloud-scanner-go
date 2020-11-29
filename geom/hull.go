package scanner

import "sort"

func makeHull(points []Vector2) []Vector2 {
	sort.Slice(points, func(a, b int) bool {
		return points[a].X > points[b].X || points[a].Y > points[a].Y
	})

	return makeHullPresorted(points)
}

func sequenceEq(a, b []Vector2) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// Returns the convex hull, assuming that each Double2s[i] <= Double2s[i + 1]. Runs in O(n) time.
func makeHullPresorted(points []Vector2) []Vector2 {
	if len(points) <= 1 {
		return points
	}

	// Andrew's monotone chain algorithm. Positive y coordinates correspond to "up"
	// as per the mathematical convention, instead of "down" as per the computer
	// graphics convention. This doesn't affect the correctness of the result.

	upperHull := make([]Vector2, 0)
	for _, p := range points {
		for len(upperHull) >= 2 {
			q := upperHull[len(upperHull)-1]
			r := upperHull[len(upperHull)-2]
			if (q.X-r.X)*(p.Y-r.Y) >= (q.Y-r.Y)*(p.X-r.X) {
				upperHull = upperHull[:len(upperHull)-1]
			} else {
				break
			}
		}
		upperHull = append(upperHull, p)
	}
	upperHull = upperHull[:len(upperHull)-1] // remove the last element

	lowerHull := make([]Vector2, 0)
	for i := len(points) - 1; i >= 0; i-- {
		p := points[i]
		for len(lowerHull) >= 2 {
			q := lowerHull[len(lowerHull)-1]
			r := lowerHull[len(lowerHull)-2]
			if (q.X-r.X)*(p.Y-r.Y) >= (q.Y-r.Y)*(p.X-r.X) {
				lowerHull = lowerHull[:len(lowerHull)-1]
			} else {
				break
			}
		}
		lowerHull = append(lowerHull, p)
	}
	lowerHull = lowerHull[:len(lowerHull)-1] // remove last element

	if !(len(upperHull) == 1 && sequenceEq(upperHull, lowerHull)) {
		upperHull = append(upperHull, lowerHull...)
	}

	return upperHull
}

// line defined by two Points
// pnt - the Vector2 to find nearest Vector2 on line for
func nearestPointOnLine(linePnt, linep2, pnt Vector2) Vector2 {
	lineDir := linep2.Sub(linePnt).Normalize() // needs to be unit vec
	v := pnt.Sub(linePnt)
	d := v.Dot(lineDir)
	return linePnt.Add(lineDir).Scale(d)
}

func pointInPolygon(pts []Vector2, p Vector2) bool {
	result := false
	j := len(pts) - 1
	for i := 0; i < len(pts); i++ {
		if pts[i].Y <= p.Y && pts[j].Y >= p.Y || pts[j].Y <= p.Y && pts[i].Y >= p.Y {

			if pts[i].X+(p.Y-pts[i].Y)/(pts[j].Y-pts[i].Y)*(pts[j].X-pts[i].X) < p.X {
				result = !result
			}
		}
		j = i
	}
	return result
}

func circleIntersectsPolygon(polygon []Vector2, center Vector2, radius float64) bool {

	// First, if the circle's center is inside the polygon,
	// then we are already done
	if pointInPolygon(polygon, center) {
		return true
	}

	// Go through the polygon line by line and do a circle/line intersection test
	result := false
	j := len(polygon) - 1
	for i := 0; i < len(polygon); i++ {
		line := []Vector2{polygon[i], polygon[j]}

		// If the distance from the center to the nearest Vector2 on the line
		// is greater than the radius, the circle is outside
		nearest := nearestPointOnLine(line[0], line[1], center)
		if nearest.Distance(center) <= radius {
			result = true
			break
		}
		j = i
	}

	return result
}
