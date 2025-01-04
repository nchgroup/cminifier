package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	ops = []string{
		"+", "-", "*", "/", "%", "++", "--",
		"+=", "-=", "*=", "/=", "%=", "=", "==", "!=",
		"&&", "||", "!", "&", "|", "^", "<<", ">>",
		"<", ">", "<=", ">=", "<<=", ">>=", "&=", "|=", "^=", ",",
		"(", ")", "{", "}", ";", "else", ":", "::", "?",
	}

	spacedOps = []string{"else"}
	unaryOps  = []string{"+", "-", "&", "!", "*"}
)

func removeEverythingBetween(start, end, line string) string {
	regex := regexp.MustCompile(regexp.QuoteMeta(start) + ".*" + regexp.QuoteMeta(end))
	return regex.ReplaceAllString(line, "")
}

func removeEverythingBefore(subs, line string) string {
	regex := regexp.MustCompile(".*" + regexp.QuoteMeta(subs))
	return regex.ReplaceAllString(line, "")
}

func removeEverythingPast(subs, line string) string {
	regex := regexp.MustCompile(regexp.QuoteMeta(subs) + ".*")
	return regex.ReplaceAllString(line, "")
}

func removeMultilineComments(lines []string) []string {
	start, end := "/*", "*/"
	inComment := false
	var newLines []string

	for _, line := range lines {
		if !inComment {
			startPos := strings.Index(line, start)
			if startPos != -1 {
				inComment = true
				endPos := strings.Index(line, end)
				if startPos < endPos {
					line = removeEverythingBetween(start, end, line)
					inComment = false
				} else {
					line = removeEverythingPast(start, line)
				}
			}
		} else {
			endPos := strings.Index(line, end)
			if endPos != -1 {
				line = removeEverythingBefore(end, line)
				inComment = false
			} else {
				line = ""
			}
		}
		newLines = append(newLines, line)
	}
	return newLines
}

func removeInlineComments(lines []string) []string {
	for i, line := range lines {
		lines[i] = removeEverythingPast("//", line)
	}
	return lines
}

func minifyOperator(op string) func(string) string {
	expr := regexp.MustCompile(` *` + regexp.QuoteMeta(op) + ` *`)
	replacement := op
	if contains(spacedOps, op) {
		replacement += " "
	}
	return func(line string) string {
		return expr.ReplaceAllString(line, replacement)
	}
}

func fixSpacedOps(text string) string {
	for _, op := range spacedOps {
		pattern := op + " {"
		replacement := op + "{"
		text = strings.ReplaceAll(text, pattern, replacement)
	}
	return text
}

func clearWhitespaceFirstPass(lines []string) []string {
	for i, line := range lines {
		lines[i] = strings.TrimSpace(strings.ReplaceAll(line, "\t", " "))
	}
	return lines
}

func reinsertPreprocessorNewlines(lines []string) []string {
	for i, line := range lines {
		if strings.HasPrefix(line, "#") {
			lines[i] += "\n"
		}
	}
	return lines
}

func fixDuplicateNewlines(file string) string {
	expr := regexp.MustCompile(`\n{2,}`)
	return expr.ReplaceAllString(file, "\n")
}

func minifySource(source string, keepNewlines, keepMultilineComments, keepInlineComments bool) string {
	lines := strings.Split(source, "\n")
	lines = clearWhitespaceFirstPass(lines)
	if !keepNewlines {
		lines = reinsertPreprocessorNewlines(lines)
	}

	for _, op := range ops {
		minifier := minifyOperator(op)
		for i, line := range lines {
			lines[i] = minifier(line)
		}
	}

	if !keepInlineComments {
		lines = removeInlineComments(lines)
	}
	if !keepMultilineComments {
		lines = removeMultilineComments(lines)
	}

	minified := strings.Join(lines, "")
	if !keepNewlines {
		minified = fixDuplicateNewlines(minified)
	}
	minified = fixSpacedOps(minified)
	return minified
}

func contains(slice []string, item string) bool {
	for _, elem := range slice {
		if elem == item {
			return true
		}
	}
	return false
}

func main() {
	inputFile := flag.String("f", "", "Path to the input file (required).")
	outputFile := flag.String("o", "", "Path to the output file. If not specified, output will be printed to the console.")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Error: You must provide an input file using the -f flag.")
		flag.Usage()
		return
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Failed to open file %s: %v\n", *inputFile, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var sourceLines []string
	for scanner.Scan() {
		sourceLines = append(sourceLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", *inputFile, err)
		return
	}

	minified := minifySource(strings.Join(sourceLines, "\n"), false, false, false)

	if *outputFile != "" {
		err = os.WriteFile(*outputFile, []byte(minified), 0644)
		if err != nil {
			fmt.Printf("Failed to write to file %s: %v\n", *outputFile, err)
			return
		}
		fmt.Printf("Minified content written to %s\n", *outputFile)
	} else {
		fmt.Println(minified)
	}
}
