package main

import (
	"bufio"
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
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	file, err := os.Open(path.Join(cwd, "data/oeis.org/stripped"))
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

func loadOEISDesc() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	file, err := os.Open(path.Join(cwd, "data/oeis.org/names"))
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

		if strings.Index(strings.ToLower(line), "decimal expansion") == -1 {
			continue
		}

		idx := strings.Index(line, " ")
		if idx == -1 {
			continue
		}

		name := strings.Trim(line[0:idx], " ")
		desc := strings.Trim(line[idx:], " ")

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

		// decimal expansion only
		if deOnly {
			// skip if the sequence ID is not in the names map
			// as of this version, the names map only contains descriptions
			// with the term `decimal expansion`
			if desc, ok = decExpNames[id]; !ok {
				continue
			}

			desc = strings.ReplaceAll(desc, "Decimal expansion of", "")
		}

		if pos >= len(seq) {
			// requested position is past the length of this sequence, skip
			continue
		}

		match := true
		for i := range digits {
			if deOnly && len(strconv.Itoa(digits[1])) > 1 {
				// if searching for a decimal expansion, skip sequences that have
				// entries longer than 1 digit
				match = false
				break
			}

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
