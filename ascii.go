package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func generateASCII(data *CollectResult) {
	// ==================== СИСТЕМНАЯ ТАБЛИЦА ====================
	fmt.Println("# system:")
	fmt.Println()

	sysHeaders := []string{"goarch", "goos", "cpu", "pkg"}
	sysContent := []string{data.Header.GOARCH, data.Header.GOOS, data.Header.CPU, data.Header.Pkg}

	var tm tableMaker
	tm.registerWidths(sysHeaders)
	tm.registerWidths(sysContent)
	tm.writeHeader(sysHeaders, true)
	tm.write(sysContent, true)

	fmt.Println()

	// ==================== ТАБЛИЦА БЕНЧМАРКОВ ====================
	fmt.Println("# benchmarks:")
	fmt.Println()

	benchHeaders := []string{"benchmark", "iter", "time/iter", "bytes alloc", "allocs"}

	tm = tableMaker{}
	tm.registerWidths(benchHeaders)
	var benchContent []string
	for _, result := range data.Results {
		benchContent = benchContent[:0]
		benchContent = append(benchContent,
			result.Name,
			strconv.Itoa(result.Iterations),
			result.NsPerOp,
		)
		if result.HasMem {
			benchContent = append(benchContent, result.BytesPerOp)
		}
		if result.HasAllocs {
			benchContent = append(benchContent, result.AllocsPerOp)
		}
		tm.registerWidths(benchContent)
	}

	tm.writeHeader(benchHeaders, true)
	for _, result := range data.Results {
		benchContent = benchContent[:0]
		benchContent = append(benchContent,
			result.Name,
			strconv.Itoa(result.Iterations),
			result.NsPerOp,
		)
		if result.HasMem {
			benchContent = append(benchContent, result.BytesPerOp)
		}
		if result.HasAllocs {
			benchContent = append(benchContent, result.AllocsPerOp)
		}
		tm.write(benchContent, true)
	}
}

type tableMaker struct {
	widths []int
	data   []byte
}

func (tm *tableMaker) registerWidths(row []string) {
	if len(row) > len(tm.widths) {
		widths := make([]int, len(row))
		copy(widths, tm.widths)
		for i := len(tm.widths); i < len(row); i++ {
			widths[i] = len(row[i])
		}
		tm.widths = widths
		return
	}

	for i, c := range row {
		tm.widths[i] = max(tm.widths[i], len(c))
	}
}

var distance = bytes.Repeat([]byte(" "), 3)

func (tm *tableMaker) writeHeader(row []string, firstLeft bool) {
	tm.data = tm.data[:0]
	for i, s := range row {
		tm.data = append(tm.data, distance...)
		tm.data = append(tm.data, tm.aligned(firstLeft && i == 0, i, s)...)
	}
	tm.data = append(tm.data, '\n')
	_, _ = os.Stdout.Write(tm.data)

	tm.data = tm.data[:0]
	for i, s := range row {
		tm.data = append(tm.data, distance...)
		tm.data = append(tm.data, tm.aligned(firstLeft && i == 0, i, strings.Repeat("-", len(s)))...)
	}
	tm.data = append(tm.data, '\n')
	_, _ = os.Stdout.Write(tm.data)
}

func (tm *tableMaker) write(row []string, firstLeft bool) {
	tm.data = tm.data[:0]
	for i, s := range row {
		tm.data = append(tm.data, distance...)
		tm.data = append(tm.data, tm.aligned(firstLeft && i == 0, i, s)...)
	}
	tm.data = append(tm.data, '\n')
	_, _ = os.Stdout.Write(tm.data)
}

func (tm *tableMaker) aligned(left bool, col int, data string) string {
	width := tm.widths[col]
	if left {
		return data + strings.Repeat(" ", width-len(data))
	}

	return strings.Repeat(" ", width-len(data)) + data
}
