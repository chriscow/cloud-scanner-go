package main

import (
	"bufio"
	"os"
	"path"
	"strings"
)

var oeisSeq map[string]string
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

	oeisSeq = make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), ",")

		name := s[0]
		oeisSeq[name] = strings.Join(s[1:], ",")
	}

	if err := scanner.Err(); err != nil {
		oeisSeq = nil
		return err
	}

	return nil
}
