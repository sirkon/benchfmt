package main

import (
	"fmt"
)

func generateMarkdown(data *CollectResult) {
	header := data.Header

	// Таблица с информацией о системе
	fmt.Println("| goos | goarch | cpu | pkg |")
	fmt.Println("|-|-|-|-|")
	fmt.Printf("| %s | %s | %s | %s |\n\n",
		header.GOOS, header.GOARCH, header.CPU, header.Pkg)

	// Заголовок результатов
	fmt.Println("## Results\n")

	// Строим заголовки для таблицы результатов
	headers := []string{"Benchmark", "Iterations", "ns/op"}
	if data.HasMem {
		headers = append(headers, "B/op")
	}
	if data.HasAllocs {
		headers = append(headers, "allocs/op")
	}

	// Выводим заголовок
	fmt.Print("|")
	for _, h := range headers {
		fmt.Printf(" %s |", h)
	}
	fmt.Println()

	// Разделитель
	fmt.Print("|")
	for range headers {
		fmt.Print(" --- |")
	}
	fmt.Println()

	// Данные
	for _, r := range data.Results {
		fmt.Printf("| %s | %v | %v", r.Name, r.Iterations, r.NsPerOp)

		if data.HasMem {
			if r.HasMem {
				fmt.Printf(" | %v", r.BytesPerOp)
			} else {
				fmt.Printf(" | —")
			}
		}

		if data.HasAllocs {
			if r.HasAllocs {
				fmt.Printf(" | %v", r.AllocsPerOp)
			} else {
				fmt.Printf(" | —")
			}
		}
		fmt.Println(" |")
	}
}
