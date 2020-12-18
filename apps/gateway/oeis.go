package main

import (
	"bufio"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

var oeisSeq map[string][]int
var oeisNames map[string]string

func loadOEIS() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	file, err := os.Open(path.Join(cwd, "data/oeis.org/oeis.org-stripped"))
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

	file, err := os.Open(path.Join(cwd, "data/oeis.org/oeis.org-names"))
	if err != nil {
		return err
	}
	defer file.Close()

	oeisNames = make(map[string]string)

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

		oeisNames[name] = desc
	}

	if err := scanner.Err(); err != nil {
		oeisSeq = nil
		return err
	}

	return nil
}

type oeisData struct {
	ID  string
	Seq []int
}

func searchOEIS(digits []int, pos int, deOnly bool) ([]oeisData, error) {
	results := make([]oeisData, 0)

	for id, seq := range oeisSeq {

		if deOnly {
			if _, ok := oeisNames[id]; !ok {
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
				ID:  id,
				Seq: seq,
			})
		}
	}

	return results, nil
}
