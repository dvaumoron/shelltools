/*
 *
 * Copyright (C) 2023  Denis Vaumoron
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dvaumoron/shelltools/pkg/common"
)

type tableBuilder = func([]int, [][]string) string

var (
	columns        []string
	simple         bool
	skipHeader     bool
	displayLineNum bool
)

func main() {
	cmd := cobra.Command{
		Use:   "jsontotable [FILE]",
		Short: "jsontotable display JSON object from FILE as a table.",
		Long: `jsontotable display JSON object from FILE as a table,
without FILE or if FILE is -, read from standard input,
without columns flag, display all attribute in sorted order (based on first object)`,
		Args: cobra.MaximumNArgs(1),
		RunE: jsonToTableWithInit,
	}

	cmdFlags := cmd.Flags()
	cmdFlags.StringSliceVarP(&columns, "columns", "c", nil, "name of the columns (comma separated)")
	cmdFlags.BoolVarP(&simple, "simple", "s", false, "simplify display (no ascii frame)")
	cmdFlags.BoolVarP(&skipHeader, "no-header", "n", false, "do not display header")
	cmdFlags.BoolVarP(&displayLineNum, "display-line-number", "l", false, "display line number")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func jsonToTableWithInit(cmd *cobra.Command, args []string) error {
	common.TrimSlice(columns)

	src, closer, err := common.GetSource(args, 0)
	if err != nil {
		return err
	}
	defer closer()

	var builder tableBuilder
	if simple {
		builder = buildLines
	} else {
		builder = func(maxColumnSizes []int, table [][]string) string {
			return buildTable(!skipHeader, maxColumnSizes, table)
		}
	}
	return jsonToTable(columns, src, skipHeader, displayLineNum, builder)
}

func jsonToTable(columns []string, src *os.File, skipHeader bool, displayLineNum bool, builder tableBuilder) error {
	initLine := initBasicLine
	if displayLineNum {
		initLine = initLineWithIndex
	}

	lineSize := 0
	var table [][]string
	scanner := bufio.NewScanner(src)
	if scanner.Scan() {
		var jsonObject map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &jsonObject); err != nil {
			return err
		}

		if len(columns) == 0 {
			columns = extractColumnNames(jsonObject, columns)
		}
		lineSize, table = initLineSizeAndTable(skipHeader, displayLineNum, columns)

		table = appendLine(table, initLine(lineSize, 0), columns, jsonObject)
	}
	for index := 1; scanner.Scan(); index++ {
		var jsonObject map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &jsonObject); err != nil {
			return err
		}

		table = appendLine(table, initLine(lineSize, index), columns, jsonObject)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return displayTable(lineSize, builder, table)
}

func initBasicLine(lineSize int, index int) []string {
	return make([]string, 0, lineSize)
}

func initLineWithIndex(lineSize int, index int) []string {
	line := make([]string, 1, lineSize)
	line[0] = strconv.Itoa(index)
	return line
}

func extractColumnNames(jsonObject map[string]any, _ []string) []string {
	names := make([]string, 0, len(jsonObject))
	for name := range jsonObject {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func initLineSizeAndTable(skipHeader bool, displayLineNum bool, columns []string) (int, [][]string) {
	lineSize := len(columns)
	if displayLineNum {
		lineSize++
	}
	if skipHeader {
		return lineSize, [][]string{}
	}
	if displayLineNum {
		header := make([]string, lineSize)
		header[0] = "#"
		copy(header[1:], columns)
		return lineSize, [][]string{header}
	}
	return lineSize, [][]string{columns}
}

func appendLine(table [][]string, line []string, columns []string, jsonObject map[string]any) [][]string {
	for _, column := range columns {
		line = append(line, common.ExtractString(jsonObject, column))
	}
	return append(table, line)
}

func displayTable(numColumns int, builder tableBuilder, table [][]string) error {
	maxColumnSizes := make([]int, numColumns)
	for _, line := range table {
		for index, value := range line {
			maxColumnSizes[index] = max(len([]rune(value)), maxColumnSizes[index])
		}
	}

	output := builder(maxColumnSizes, table)
	_, err := os.Stdout.WriteString(output)
	return err
}

func buildTable(displayHeader bool, maxColumnSizes []int, table [][]string) string {
	interline := buildInterline(maxColumnSizes)

	var outputBuilder strings.Builder
	outputBuilder.WriteString(interline)
	tableSize := len(table)
	writeTableLine(&outputBuilder, table[0], maxColumnSizes)
	if displayHeader {
		outputBuilder.WriteString(interline)
	}
	for index := 1; index < tableSize; index++ {
		writeTableLine(&outputBuilder, table[index], maxColumnSizes)
	}
	outputBuilder.WriteString(interline)
	return outputBuilder.String()
}

func buildInterline(maxColumnSizes []int) string {
	var builder strings.Builder
	builder.WriteByte('+')
	for _, maxColumnSize := range maxColumnSizes {
		builder.WriteByte('-')
		for counter := 0; counter < maxColumnSize; counter++ {
			builder.WriteByte('-')
		}
		builder.WriteString("-+")
	}
	builder.WriteByte('\n')
	return builder.String()
}

func writeTableLine(outputBuilder *strings.Builder, line []string, maxColumnSizes []int) {
	outputBuilder.WriteByte('|')
	for index, value := range line {
		outputBuilder.WriteByte(' ')
		outputBuilder.WriteString(value)
		for counter := len([]rune(value)); counter < maxColumnSizes[index]; counter++ {
			outputBuilder.WriteByte(' ')
		}
		outputBuilder.WriteString(" |")
	}
	outputBuilder.WriteByte('\n')
}

func buildLines(maxColumnSizes []int, table [][]string) string {
	var outputBuilder strings.Builder
	for _, line := range table {
		for index, value := range line {
			outputBuilder.WriteString(value)
			for counter := len([]rune(value)); counter < maxColumnSizes[index]; counter++ {
				outputBuilder.WriteByte(' ')
			}
			outputBuilder.WriteByte(' ')
		}
		outputBuilder.WriteByte('\n')
	}
	return outputBuilder.String()
}
