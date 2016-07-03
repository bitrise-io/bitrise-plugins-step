package utils

import (
	"bufio"
	"math"
	"strings"
)

// IndentTextWithMaxLength ...
func IndentTextWithMaxLength(text string, indent int, maxLineCharWidth int) string {
	formattedText := ""
	indentStr := strings.Repeat(" ", indent)
	maxCharPerLine := maxLineCharWidth - indent

	addLine := func(line string) {
		if formattedText != "" {
			formattedText = formattedText + "\n"
		}
		formattedText = formattedText + indentStr + line
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		lineLength := len(line)
		if lineLength > maxCharPerLine {
			lineCnt := math.Ceil(float64(lineLength) / float64(maxCharPerLine))
			for i := 0; i < int(lineCnt); i++ {
				startIdx := i * maxCharPerLine
				endIdx := startIdx + maxCharPerLine
				if endIdx > lineLength {
					endIdx = lineLength
				}
				addLine(line[startIdx:endIdx])
			}
		} else {
			addLine(line)
		}
	}

	return formattedText
}
