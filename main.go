package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jaypipes/ghw"
)

const bucketCount = 3600

type latticePoints struct {
	Points []float64 `json:"Points"`
}

type point struct {
	x, y float64
}

func pointIndex(name string) (int, error) {
	tok := strings.Split(name, "-")
	vals := strings.Split(tok[1], ".")
	return strconv.Atoi(vals[0])
}

func loadPoints(path string) []point {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatalln(err)
	}

	s := new(sync.WaitGroup)
	m := new(sync.Mutex)
	points := make([]point, 0, 1024*1024)

	s.Add(len(files))
	for _, f := range files {
		go func(name string) {
			const pointCount = 8192

			data, err := ioutil.ReadFile(path + name)
			if err != nil {
				log.Fatalln(err)
			}

			var lp latticePoints
			if err := json.Unmarshal(data, &lp); err != nil {
				log.Fatalln(err)
			}

			// Each filename has a number that is the starting index
			pts := make([]point, 0, pointCount)
			for i := 0; i < len(lp.Points); i += 2 {
				pt := point{
					x: lp.Points[i],
					y: lp.Points[i+1],
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
	return points
}

func loadZeros(path string, max int) []float64 {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	zeros := make([]float64, 0, 1024)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		value, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			log.Fatalln(err)
		}

		zeros = append(zeros, value)
		max--
		if max == 0 {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return zeros
}

func randOrigins(min, max float64, center point, count int) []point {
	res := make([]point, count)

	for i := range res {
		res[i] = point{
			x: min + rand.Float64()*(max-min) + center.x,
			y: min + rand.Float64()*(max-min) + center.y,
		}
	}
	return res
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

func filterLattice(origin point, radius float64, lattice []point, zeros []float64, distanceLimit float64) []point {

	maxZero := max(zeros)
	points := make([]point, 0, len(lattice))

	// Calculate the radius based on ZLines selected and DistanceLimit
	// Double it because the origin can be at the edge of this
	r := math.Sqrt((radius+maxZero)*(radius+maxZero) + distanceLimit*distanceLimit)

	for _, pt := range lattice {
		// if its in the radius, copy the point and move the index
		distance := math.Sqrt((pt.x-origin.x)*(pt.x-origin.x) + (pt.y-origin.y)*(pt.y-origin.y))
		if math.Abs(distance) <= r {
			points = append(points, pt)
		}
	}

	return points
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

func allAngles(lattice, origin point, zero float64) (theta1, theta2 float64) {
	xSq := (lattice.x - origin.x) * (lattice.x - origin.x)
	ySq := (lattice.y - origin.y) * (lattice.y - origin.y)
	zSq := zero * zero

	distance := math.Sqrt(xSq + ySq - zSq)

	if math.IsNaN(distance) {
		return math.NaN(), math.NaN()
	}

	rad2deg := 180 / math.Pi
	theta1 = wrapDegrees(rad2deg * 2 * math.Atan2(lattice.y-origin.y+distance, lattice.x-origin.x+zero))
	theta2 = wrapDegrees(rad2deg * 2 * math.Atan2(lattice.y-origin.y-distance, lattice.x-origin.x+zero))

	return
}

func calculate(center point, radius float64, scans int, lattice []point, zeros []float64, wg *sync.WaitGroup) {
	buckets := make([]int, 3600, 3600)
	degPerBucket := 360.0 / float64(len(buckets))

	origins := randOrigins(-radius, radius, center, scans)

	for _, point := range lattice {
		for _, zero := range zeros {
			for _, origin := range origins {
				theta1, theta2 := allAngles(point, origin, zero)

				if math.IsNaN(theta1) || math.IsNaN(theta2) {
					continue
				}

				b1 := int(math.Floor(theta1 / degPerBucket))
				b2 := int(math.Floor(theta2 / degPerBucket))
				// log.Println(theta1, theta2, b1, b2)

				if b1 >= len(buckets) || b2 >= len(buckets) || b1 < 0 || b2 < 0 {
					log.Fatalln("bucket", b1, "out of range:", len(buckets), lattice, origin, zero, theta1, theta2)
				}
				buckets[b1]++
				if b1 != b2 {
					buckets[b2]++
				}
			}
		}
	}

	wg.Done()
}

func main() {
	path := "../lattices/Pinwheel/Vertices/8192/points/"
	lattice := loadPoints(path)

	path = "../zeros/primes.txt"
	zeros := loadZeros(path, 25)

	threads := 8

	log.Println("os", runtime.GOOS)
	if runtime.GOOS != "darwin" {
		threads = 0
		cpu, err := ghw.CPU()
		if err != nil {
			fmt.Printf("Error getting CPU info: %v", err)
		}

		for _, proc := range cpu.Processors {
			fmt.Printf(" %v\n", proc)
			for _, core := range proc.Cores {
				fmt.Printf("  %v\n", core)
			}
			if len(proc.Capabilities) > 0 {
				// pretty-print the (large) block of capability strings into rows
				// of 6 capability strings
				rows := int(math.Ceil(float64(len(proc.Capabilities)) / float64(6)))
				for row := 1; row < rows; row = row + 1 {
					rowStart := (row * 6) - 1
					rowEnd := int(math.Min(float64(rowStart+6), float64(len(proc.Capabilities))))
					rowElems := proc.Capabilities[rowStart:rowEnd]
					capStr := strings.Join(rowElems, " ")
					if row == 1 {
						fmt.Printf("  capabilities: [%s\n", capStr)
					} else if rowEnd < len(proc.Capabilities) {
						fmt.Printf("                 %s\n", capStr)
					} else {
						fmt.Printf("                 %s]\n", capStr)
					}
				}
			}
		}

		threads = int(cpu.TotalThreads)
		log.Println("total cores:", cpu.TotalCores, "total threads:", cpu.TotalThreads)
	}

	log.Println("threads:", threads)
	scansPerThread := 300

	center := point{x: 0, y: 0}
	radius := 1.0
	lattice = filterLattice(center, radius, lattice, zeros, 1)

	wg := new(sync.WaitGroup)
	wg.Add(threads)

	start := time.Now()
	for i := 0; i < threads; i++ {
		go calculate(center, radius, scansPerThread, lattice, zeros, wg)
	}

	wg.Wait()
	elapsed := time.Since(start)
	scansPerSec := float64(threads*scansPerThread) / elapsed.Seconds()
	log.Println("elapsed", elapsed, scansPerSec, "scans/sec")
}
