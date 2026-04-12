// vs.go
package main

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type ComparedResult struct {
	CommonName    string
	Value1        string
	Value2        string
	Ratio         float64
	HasValue1     bool
	HasValue2     bool
	OriginalOrder int
}

func generateVS(data *CollectResult, vsSpec string, format string) {
	// Парсим спецификацию: "BinLog:ZeroLog" или "std:sfk"
	parts := strings.Split(vsSpec, ":")
	if len(parts) != 2 {
		fmt.Fprintf(os.Stderr, "Invalid --vs format. Expected 'name1:name2', got '%s'\n", vsSpec)
		os.Exit(1)
	}

	part0 := parts[0]
	part1 := parts[1]

	// Группируем результаты по общему имени после замены
	group1 := make(map[string]BenchResult) // commonName -> result
	group2 := make(map[string]BenchResult)
	orderMap := make(map[string]int) // commonName -> порядок из part[0]

	// Собираем все возможные commonName
	allCommonNames := make(map[string]bool)
	orderCounter := 0

	for _, r := range data.Results {
		// Проверяем, содержит ли имя part0
		if strings.Contains(r.Name, part0) {
			// Заменяем part0 на "VS" для наглядности
			commonName := strings.Replace(r.Name, part0, "", 1)
			group1[commonName] = r
			allCommonNames[commonName] = true
			// Сохраняем порядок только для part[0]
			if _, exists := orderMap[commonName]; !exists {
				orderMap[commonName] = orderCounter
				orderCounter++
			}
		}

		// Проверяем, содержит ли имя part1
		if strings.Contains(r.Name, part1) {
			// Заменяем part1 на "VS" для наглядности
			commonName := strings.Replace(r.Name, part1, "", 1)
			group2[commonName] = r
			allCommonNames[commonName] = true
			// Не обновляем порядок для part[1], только если его еще нет
			if _, exists := orderMap[commonName]; !exists {
				orderMap[commonName] = orderCounter
				orderCounter++
			}
		}
	}

	if len(allCommonNames) == 0 {
		fmt.Fprintf(os.Stderr, "No benchmarks found matching '%s' or '%s'\n", part0, part1)
		os.Exit(1)
	}

	// Находим общий префикс для всех commonName
	commonPrefix := findCommonPrefix(allCommonNames)

	// Сортируем по порядку из part[0]
	type nameWithOrder struct {
		fullName    string
		trimmedName string
		order       int
	}

	var sortedNames []nameWithOrder
	for fullName := range allCommonNames {
		trimmedName := strings.TrimPrefix(fullName, commonPrefix)
		// Убираем из trimmedName все что не буква или число.
		runes := []rune(trimmedName)
		for len(runes) > 0 {
			r := runes[0]
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				break
			}
			runes = runes[1:]
		}
		slices.Reverse(runes)
		for len(runes) > 0 {
			r := runes[0]
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				break
			}
			runes = runes[1:]
		}
		slices.Reverse(runes)
		trimmedName = string(runes)
		// Если trimmedName пустой, используем что-то осмысленное
		if trimmedName == "" {
			trimmedName = fullName
		}
		sortedNames = append(sortedNames, nameWithOrder{
			fullName:    fullName,
			trimmedName: trimmedName,
			order:       orderMap[fullName],
		})
	}

	// Сортируем по order
	sort.Slice(sortedNames, func(i, j int) bool {
		return sortedNames[i].order < sortedNames[j].order
	})

	// Формируем результаты для сравнения
	var comparisons []ComparedResult
	for _, item := range sortedNames {
		comp := ComparedResult{
			CommonName:    item.trimmedName,
			OriginalOrder: item.order,
		}

		if r1, ok := group1[item.fullName]; ok {
			comp.Value1 = r1.NsPerOp
			comp.HasValue1 = true
		} else {
			comp.Value1 = "—"
		}

		if r2, ok := group2[item.fullName]; ok {
			comp.Value2 = r2.NsPerOp
			comp.HasValue2 = true
		} else {
			comp.Value2 = "—"
		}

		// Вычисляем отношение: part1 / part0
		if comp.HasValue1 && comp.HasValue2 {
			val1 := parseNsPerOp(comp.Value1)
			val2 := parseNsPerOp(comp.Value2)
			if val1 > 0 && val2 > 0 {
				comp.Ratio = val2 / val1
			}
		}

		comparisons = append(comparisons, comp)
	}

	// Выводим в зависимости от формата
	if format == "md" {
		generateVSMarkdown(data, part0, part1, comparisons)
	} else {
		generateVSASCII(data, part0, part1, comparisons)
	}
}

func generateVSASCII(data *CollectResult, part0, part1 string, comparisons []ComparedResult) {
	// Выводим системную информацию
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

	// Выводим сравнение
	fmt.Printf("# comparison: %s vs %s\n", part0, part1)
	fmt.Println()

	// Заголовки таблицы
	headers := []string{"test", part0, part1, fmt.Sprintf("ratio (%s/%s)", part1, part0)}

	tm = tableMaker{}
	tm.registerWidths(headers)

	// Регистрируем ширины для всех строк
	for _, comp := range comparisons {
		row := []string{
			comp.CommonName,
			comp.Value1,
			comp.Value2,
			formatRatio(comp.Ratio),
		}
		tm.registerWidths(row)
	}

	// Выводим таблицу
	tm.writeHeader(headers, true)
	for _, comp := range comparisons {
		row := []string{
			comp.CommonName,
			comp.Value1,
			comp.Value2,
			formatRatio(comp.Ratio),
		}
		tm.write(row, true)
	}
}

func generateVSMarkdown(data *CollectResult, part0, part1 string, comparisons []ComparedResult) {
	// Системная информация
	fmt.Println("**System**")
	fmt.Println()

	fmt.Println("| goos | goarch | cpu | pkg |")
	fmt.Println("|-|-|-|-|")
	fmt.Printf("| %s | %s | %s | %s |\n\n",
		data.Header.GOOS, data.Header.GOARCH, data.Header.CPU, data.Header.Pkg)

	// Заголовок сравнения
	fmt.Printf("**Comparison: %s vs %s**\n\n", part0, part1)

	// Заголовки таблицы
	headers := []string{"Test", part0, part1, "Ratio (2nd/1st)"}

	fmt.Print("|")
	for _, h := range headers {
		fmt.Printf(" %s |", h)
	}
	fmt.Println()

	fmt.Print("|")
	for range headers {
		fmt.Print(" --- |")
	}
	fmt.Println()

	// Данные
	for _, comp := range comparisons {
		fmt.Printf("| %s | %s | %s | %s |\n",
			comp.CommonName,
			comp.Value1,
			comp.Value2,
			formatRatio(comp.Ratio),
		)
	}
}

// findCommonPrefix находит общий префикс для всех строк в map
func findCommonPrefix(names map[string]bool) string {
	if len(names) == 0 {
		return ""
	}

	// Получаем первый элемент как baseline
	var first string
	for name := range names {
		first = name
		break
	}

	// Ищем общий префикс со всеми остальными
	for name := range names {
		for i := 0; i < len(first) && i < len(name); i++ {
			if first[i] != name[i] {
				first = first[:i]
				break
			}
		}
	}

	return first
}

// parseNsPerOp парсит строку типа "43.64 ns/op" и возвращает число наносекунд
func parseNsPerOp(s string) float64 {
	// Убираем " ns/op" в конце
	s = strings.TrimSuffix(s, " ns/op")
	s = strings.TrimSpace(s)

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

func formatRatio(ratio float64) string {
	if ratio == 0 {
		return "—"
	}

	// ratio = part1/part0
	return fmt.Sprintf("%.2fx", ratio)
}
