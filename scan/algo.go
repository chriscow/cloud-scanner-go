package scan

import (
	"log"
	"math"
	"math/rand"
	"github.com/chriscow/cloud-scanner-go/geom"
)

func randOrigins(min, max float64, center geom.Vector2, count int) []geom.Vector2 {
	origins := make([]geom.Vector2, count)
	for i := range origins {
		origins[i] = geom.Vector2{
			X: min + rand.Float64()*(max-min) + center.X,
			Y: min + rand.Float64()*(max-min) + center.Y,
		}
	}
	return origins
}

func wrapDegrees(deg float64) float64 {
	for deg > 360 {
		deg -= 360
	}

	for deg < 0 {
		deg += 360
	}

	return deg
}

func allAngles(lattice, origin geom.Vector2, zero, limit float64) (theta1, theta2 float64) {
	xSq := (lattice.X - origin.X) * (lattice.X - origin.X)
	ySq := (lattice.Y - origin.Y) * (lattice.Y - origin.Y)
	zSq := zero * zero

	distance := math.Sqrt(xSq + ySq - zSq)

	if math.IsNaN(distance) || distance > limit {
		return math.NaN(), math.NaN()
	}

	rad2deg := 180 / math.Pi
	theta1 = wrapDegrees(rad2deg * 2 * math.Atan2(lattice.Y-origin.Y+distance, lattice.X-origin.X+zero))
	theta2 = wrapDegrees(rad2deg * 2 * math.Atan2(lattice.Y-origin.Y-distance, lattice.X-origin.X+zero))

	return
}

// calculate a single result
func calculate(origin geom.Vector2, lattice []geom.Vector2, zeros []float64, latticeParams interface{}, limit float64, bucketCount int) [][]int {
	buckets := make([][]int, bucketCount)
	for i := range buckets {
		buckets[i] = make([]int, len(zeros))
	}

	degPerBucket := 360.0 / float64(len(buckets))

	for _, Vector2 := range lattice {
		for i, zero := range zeros {
			theta1, theta2 := allAngles(Vector2, origin, zero, limit)

			if math.IsNaN(theta1) || math.IsNaN(theta2) {
				continue
			}

			b1 := int(math.Floor(theta1 / degPerBucket))
			b2 := int(math.Floor(theta2 / degPerBucket))
			// log.Println(theta1, theta2, b1, b2)

			if b1 >= len(buckets) || b2 >= len(buckets) || b1 < 0 || b2 < 0 {
				log.Fatalln("bucket", b1, "out of range:", len(buckets), lattice, origin, zero, theta1, theta2)
			}
			buckets[b1][i] = 1
			if b1 != b2 {
				buckets[b2][i] = 1
			}
		}
	}

	return buckets
}

func calculateTest(origin geom.Vector2, lattice []geom.Vector2, zeros geom.Zeros, latticeParams interface{}, limit float64, bucketCount int) [][]int {
	buckets := make([][]int, bucketCount)
	for i := range buckets {
		buckets[i] = make([]int, zeros.Count)
	}

	degPerBucket := 360.0 / float64(len(buckets))

	for _, Vector2 := range lattice {
		for i, zero := range zeros.Values {
			theta1, theta2 := allAngles(Vector2, origin, zero, limit)

			if math.IsNaN(theta1) || math.IsNaN(theta2) {
				continue
			}

			b1 := int(math.Floor(theta1 / degPerBucket))
			b2 := int(math.Floor(theta2 / degPerBucket))
			// log.Println(theta1, theta2, b1, b2)

			if b1 >= len(buckets) || b2 >= len(buckets) || b1 < 0 || b2 < 0 {
				log.Fatalln("bucket", b1, "out of range:", len(buckets), lattice, origin, zero, theta1, theta2)
			}
			buckets[b1][i] = 1
			if b1 != b2 {
				buckets[b2][i] = 1
			}
		}
	}

	return buckets
}
