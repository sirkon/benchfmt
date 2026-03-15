package main

import (
	"bufio"
	"fmt"
	"io"
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

func Collect(src io.Reader) (*CollectResult, error) {
	var res CollectResult
	scanner := bufio.NewScanner(src)

	var void string
	var maxBenchNameLength int
	for scanner.Scan() {
		line := scanner.Text()
		// Заменяем табы и множественные пробелы
		fields := strings.Fields(strings.ReplaceAll(line, "\t", " "))
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "goos:":
			res.Header.GOOS = strings.Join(fields[1:], " ")
		case "goarch:":
			res.Header.GOARCH = strings.Join(fields[1:], " ")
		case "pkg:":
			res.Header.Pkg = strings.Join(fields[1:], " ")
		case "cpu:":
			res.Header.CPU = strings.Join(fields[1:], " ")
		default:
			if strings.HasPrefix(fields[0], "Benchmark") {
				if len(fields[0]) > maxBenchNameLength {
					maxBenchNameLength = len(fields[0])
				}
				if maxBenchNameLength > len(void) {
					void = strings.Repeat(" ", maxBenchNameLength)
				}
				// Ожидаем минимум 4 поля: имя, итерации, значение, ns/op
				fmt.Printf("\r\033[3m%s\r%s\033[0m", void, fields[0])

				if len(fields) < 4 {
					continue
				}
				name := fields[0]
				iter, _ := strconv.Atoi(fields[1])
				r := BenchResult{
					Name:       name,
					Iterations: iter,
					NsPerOp:    fields[2] + " " + fields[3],
				}
				idx := 4
				// Парсим память (B/op)
				if idx+1 < len(fields) && fields[idx+1] == "B/op" {
					r.BytesPerOp = fields[idx] + " " + fields[idx+1]
					r.HasMem = true
					idx += 2
				}
				// Парсим аллокации (allocs/op)
				if idx+1 < len(fields) && fields[idx+1] == "allocs/op" {
					r.AllocsPerOp = fields[idx] + " " + fields[idx+1]
					r.HasAllocs = true
					idx += 2
				}
				res.Results = append(res.Results, r)
			}
			// остальные строки игнорируем
		}
	}
	fmt.Printf("\r%s\r", strings.Repeat(" ", maxBenchNameLength))

	// Устанавливаем флаги наличия метрик и отсекаем суффикс -\d+ если он есть на всех результатах.
	hasCommonSuffix := true
	var commonSuffix string
	for _, r := range res.Results {
		if hasCommonSuffix && matchSuffix.MatchString(r.Name) {
			if !fixedSuffix {
				fixedSuffix = true
				pos := strings.LastIndex(r.Name, "-")
				commonSuffix = r.Name[pos:]
				matchSuffix = regexp.MustCompile(`^.*-` + r.Name[pos+1:] + `*$`)
			}
		} else {
			hasCommonSuffix = false
		}
		if r.HasMem {
			res.HasMem = true
		}
		if r.HasAllocs {
			res.HasAllocs = true
		}
	}

	if hasCommonSuffix {
		for i := range res.Results {
			res.Results[i].Name = res.Results[i].Name[:len(res.Results[i].Name)-len(commonSuffix)]
		}
	}

	return &res, scanner.Err()
}

var matchSuffix = regexp.MustCompile(`^.*-\d*$`)
var fixedSuffix bool
