package main

// ScanResult holds the data from a single scan and is serialized with MessagePack
type scanResult struct {
	Origin        Point
	DistanceLimit float64
	BestTheta     float64
	ZeroType      ZeroType
	Scalar        float64
	ZerosCount    int
	ZerosHit      int
	AvgParity     float64
	LatticeParams interface{}
	Score         float64 `msgpack:"ignore"`
}
