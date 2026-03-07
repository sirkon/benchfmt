package main

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type BenchHeader struct {
	GOOS   string
	GOARCH string
	Pkg    string
	CPU    string
}

type BenchResult struct {
	Name        string
	Iterations  int
	NsPerOp     string // Теперь строка с числом и единицей измерения
	BytesPerOp  string // Строка с числом и B/op или пустая
	AllocsPerOp string // Строка с числом и allocs/op или пустая
	HasMem      bool
	HasAllocs   bool
}

type CollectResult struct {
	Header  BenchHeader
	Results []BenchResult

	HasMem    bool
	HasAllocs bool
}

func Collect() (*CollectResult, error) {
	var res CollectResult

	// Регулярки
	headerRe := regexp.MustCompile(`^(\w+):\s+(.+)$`)
	// Новая регулярка, которая захватывает всё вместе
	benchRe := regexp.MustCompile(`^(Benchmark\w+)-\d+\s+(\d+)\s+([\d.]+\s+ns/op)(?:\s+(\d+\s+B/op))?(?:\s+(\d+\s+allocs/op))?$`)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Парсим заголовок
		if matches := headerRe.FindStringSubmatch(line); matches != nil {
			switch matches[1] {
			case "goos":
				res.Header.GOOS = matches[2]
			case "goarch":
				res.Header.GOARCH = matches[2]
			case "pkg":
				res.Header.Pkg = matches[2]
			case "cpu":
				res.Header.CPU = matches[2]
			}
			continue
		}

		// Парсим бенчмарки
		if matches := benchRe.FindStringSubmatch(line); matches != nil {
			r := BenchResult{
				Name:       matches[1],
				Iterations: parseInt(matches[2]),
				NsPerOp:    matches[3], // Уже с "ns/op"
			}

			// Bytes alloc
			if len(matches) > 4 && matches[4] != "" {
				r.BytesPerOp = matches[4] // Уже с "B/op"
				r.HasMem = true
				res.HasMem = true
			}

			// Allocs
			if len(matches) > 5 && matches[5] != "" {
				r.AllocsPerOp = matches[5] // Уже с "allocs/op"
				r.HasAllocs = true
				res.HasAllocs = true
			}

			res.Results = append(res.Results, r)
		}
	}

	if err := scanner.Err(); err != nil {
		return &res, err
	}

	return &res, nil
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
