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
	"strconv"
	"strings"

	"github.com/dvaumoron/shelltools/common"
	"github.com/spf13/cobra"
)

type tableBuilder = func([]int, [][]string) string

var columns []string
var simple bool
var skipHeader bool
var hideLineNum bool

func main() {
	cmd := cobra.Command{
		Use:   "jsontotable [FILE]",
		Short: "jsontotable display JSON object from FILE as a table.",
		Long: `jsontotable display JSON object from FILE as a table,
without FILE or if FILE is -, read from standard input`,
		Args: cobra.RangeArgs(1, 2),
		RunE: jsonToTableWithInit,
	}

	cmdFlags := cmd.Flags()
	cmdFlags.StringSliceVarP(&columns, "columns", "c", nil, "name of the columns (comma separated)")
	cmdFlags.BoolVarP(&simple, "simple", "s", false, "simplify display (no ascii frame)")
	cmdFlags.BoolVarP(&skipHeader, "no-header", "n", false, "do not display header")
	cmdFlags.BoolVarP(&hideLineNum, "hide-line-number", "h", false, "hide line number")

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

	builder := buildTable
	if simple {
		builder = buildLines
	}
	return jsonToTable(columns, src, skipHeader, !hideLineNum, builder)
}

func jsonToTable(columns []string, src *os.File, skipHeader bool, displayLineNum bool, builder tableBuilder) error {
	lineSize, table := initLineSizeAndTable(skipHeader, displayLineNum, columns)
	initLine := initBasicLine
	if displayLineNum {
		initLine = initLineWithIndex
	}

	scanner := bufio.NewScanner(src)
	for index := 0; scanner.Scan(); index++ {
		var jsonObject map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &jsonObject); err != nil {
			return err
		}
		line := initLine(lineSize, index)
		for _, column := range columns {
			line = append(line, common.ExtractString(jsonObject, column))
		}
		table = append(table, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return display(lineSize, builder, table)
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

func initBasicLine(lineSize int, index int) []string {
	return make([]string, 0, lineSize)
}

func initLineWithIndex(lineSize int, index int) []string {
	line := make([]string, 1, lineSize)
	line[0] = strconv.Itoa(index)
	return line
}

func display(numColumns int, builder tableBuilder, table [][]string) error {
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

func buildTable(maxColumnSizes []int, table [][]string) string {
	interline := buildInterline(maxColumnSizes)

	var outputBuilder strings.Builder
	outputBuilder.WriteString(interline)
	for index, line := range table {
		if index == 1 {
			outputBuilder.WriteString(interline)
		}
		outputBuilder.WriteByte('|')
		for index2, value := range line {
			outputBuilder.WriteByte(' ')
			outputBuilder.WriteString(value)
			for counter := len([]rune(value)); counter < maxColumnSizes[index2]; counter++ {
				outputBuilder.WriteByte(' ')
			}
			outputBuilder.WriteString(" |")
		}
		outputBuilder.WriteByte('\n')
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
