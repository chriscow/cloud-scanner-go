package main

import (
	"github.com/joho/godotenv"
)

const bucketCount = 3600

func checkEnv() {
	godotenv.Load()

}

func main() {
	checkEnv()

	sampleScan(25, 100) // zero count, scans per thread
}
