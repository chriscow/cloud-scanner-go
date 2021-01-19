package main

import (
	"bufio"
	"errors"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

// Databases are available here: https://oeis.org/wiki/Welcome#Compressed_Versions

var oeisSeq map[string][]int
var decExpNames map[string]string

func loadOEIS() error {
	oeisPath := path.Join(os.Getenv("APP_DATA"), "oeis.org")

	file, err := os.Open(path.Join(oeisPath, "stripped"))
	if err != nil {
		return err
	}
	defer file.Close()

	oeisSeq = make(map[string][]int)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// skip comments
		if line[0] == '#' {
			continue
		}

		if line[0] != 'A' {
			log.Println("Unexpected start of line. Skipping", line)
			continue
		}

		s := strings.Split(line, ",")

		nums := make([]int, 0, 256)
		name := strings.Trim(s[0], " ")

		for i := 1; i < len(s); i++ {
			num, err := strconv.Atoi(s[i])
			if err != nil {
				continue
			}
			nums = append(nums, num)
		}

		oeisSeq[name] = nums
	}

	if err := scanner.Err(); err != nil {
		oeisSeq = nil
		return err
	}

	return nil
}

func loadOEISDecExp() error {
	if len(oeisSeq) == 0 {
		return errors.New("you must call loadOEIS() first")
	}

	oeisPath := path.Join(os.Getenv("APP_DATA"), "oeis.org")

	file, err := os.Open(path.Join(oeisPath, "names"))
	if err != nil {
		return err
	}
	defer file.Close()

	decExpNames = make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// skip comments
		if line[0] == '#' {
			continue
		}

		if line[0] != 'A' {
			log.Println("Unexpected start of line. Skipping", line)
			continue
		}

		descIdx := strings.Index(strings.ToLower(line), "decimal expansion of")
		if descIdx == -1 {
			continue
		}

		descIdx += len("decimal expansion of ")

		idx := strings.Index(line, " ")
		if idx == -1 {
			continue
		}

		name := strings.Trim(line[0:idx], " ")

		// only load decimal expansion entries where all digits are single digit
		// (<= 9)
		seq := oeisSeq[name]
		keep := func(seq []int) bool {
			for i := range seq {
				if seq[i] > 9 {
					return false
				}
			}
			return true
		}(seq)

		if !keep {
			continue
		}

		desc := strings.Trim(line[descIdx:], " ")
		desc = strings.TrimRight(desc, ".")

		decExpNames[name] = desc
	}

	if err := scanner.Err(); err != nil {
		oeisSeq = nil
		return err
	}

	return nil
}

type oeisData struct {
	ID   string
	Desc string
	Seq  []int
}

func searchOEIS(digits []int, pos int, deOnly bool) ([]oeisData, error) {
	results := make([]oeisData, 0)
	var ok bool
	var desc string
	for id, seq := range oeisSeq {

		// decimal expansion only.
		// TODO: could be optimized to only search sequences
		// in the decExpNames map but this is fast enough for now
		if deOnly {
			// skip if the sequence ID is not in the names map
			// as of this version, the names map only contains descriptions
			// with the term `decimal expansion`
			if desc, ok = decExpNames[id]; !ok {
				continue
			}
		}

		if pos >= len(seq) {
			// requested position is past the length of this sequence, skip
			continue
		}

		match := true
		for i := range digits {
			if pos+i >= len(seq) || seq[pos+i] != digits[i] {
				match = false
				break
			}
		}

		if match {
			results = append(results, oeisData{
				ID:   id,
				Desc: desc,
				Seq:  seq,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		id1, _ := strconv.Atoi(results[i].ID[1:])
		id2, _ := strconv.Atoi(results[j].ID[1:])

		return id1 < id2
	})

	return results, nil
}
