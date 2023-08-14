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
)

func main() {
	cmd := cobra.Command{
		Use:   "jsontotable COLUMNS [FILE]",
		Short: "jsontotable display JSON object from FILE as a table.",
		Long:  "todo",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  jsonToTableWithInit,
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func jsonToTableWithInit(cmd *cobra.Command, args []string) error {
	columns := spaceSplitter(args[0]) // TODO make this optional based on order of attribute in first object
	// TODO add an optional flag to sort on a column (with an int converted version)
	// TODO add an optional flag to limit number of line to output

	src := os.Stdin
	if len(args) != 1 {
		if filePath := args[1]; filePath != "-" {
			src, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer src.Close()
		}
	}
	return jsonToTable(columns, src)
}

func spaceSplitter(rawValues string) []string {
	splitted := strings.Split(rawValues, " ")
	values := make([]string, 0, len(splitted))
	for _, value := range splitted {
		if value != "" {
			values = append(values, value)
		}
	}
	return slices.Clip(values)
}

func jsonToTable(columns []string, src *os.File) error {
	scanner := bufio.NewScanner(src)

	capacity := len(columns) + 1
	header := make([]string, capacity)
	header[0] = "#"
	copy(header[1:], columns)

	table := [][]string{header}
	for index := 0; scanner.Scan(); index++ {
		var jsonObject map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &jsonObject); err != nil {
			return err
		}
		line := make([]string, 1, capacity)
		line[0] = strconv.Itoa(index)
		for _, column := range columns {
			line = append(line, fmt.Sprint(jsonObject[column]))
		}
		table = append(table, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return displayTable(capacity, table)
}

func displayTable(numColumns int, table [][]string) error {
	maxColumnSizes := make([]int, numColumns)
	for _, line := range table {
		for index, value := range line {
			maxColumnSizes[index] = max(len([]rune(value)), maxColumnSizes[index])
		}
	}
	interline := buildInterline(maxColumnSizes)

	var outputBuilder strings.Builder
	outputBuilder.WriteString(interline)
	for index, line := range table {
		if index == 1 {
			outputBuilder.WriteString(interline)
		}
		outputBuilder.WriteByte('+')
		for index2, value := range line {
			outputBuilder.WriteByte(' ')
			outputBuilder.WriteString(value)
			for counter := len([]rune(value)); counter < maxColumnSizes[index2]; counter++ {
				outputBuilder.WriteByte(' ')
			}
			outputBuilder.WriteString(" +")
		}
		outputBuilder.WriteByte('\n')
	}
	outputBuilder.WriteString(interline)

	_, err := os.Stdout.WriteString(outputBuilder.String())
	return err
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
