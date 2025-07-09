package log

import (
	"fmt"
	"strings"
)

// 문자열을 지정한 길이만큼 양쪽에 공백을 추가해 중앙정렬
func padCenter(s string, n int) string {
	if len(s) >= n {
		return s
	}
	left := (n - len(s)) / 2
	right := n - len(s) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// 중앙정렬 마크다운 테이블 출력
func PrintMarkdownTable(headers []string, rows [][]string) {
	numCols := len(headers)
	colWidths := make([]int, numCols)
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, v := range row {
			if l := len(v); l > colWidths[i] {
				colWidths[i] = l
			}
		}
	}
	head := "|"
	sep := "|"
	for i, h := range headers {
		pad := colWidths[i]
		head += " " + padCenter(h, pad) + " |"
		sep += ":" + strings.Repeat("-", pad) + ":|"
	}
	fmt.Println(head)
	fmt.Println(sep)
	for _, row := range rows {
		fmt.Print("|")
		for i, v := range row {
			pad := colWidths[i]
			fmt.Printf(" %s |", padCenter(v, pad))
		}
		fmt.Println()
	}
}
