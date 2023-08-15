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
	"strings"

	"github.com/dvaumoron/shelltools/common"
	"github.com/spf13/cobra"
)

func main() {
	cmd := cobra.Command{
		Use:   "linetojson COLUMNS [FILE]",
		Short: "linetojson convert each line from FILE in a JSON object.",
		Long:  "todo",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  lineToJsonWithInit,
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func lineToJsonWithInit(cmd *cobra.Command, args []string) error {
	columns := common.SpaceSplitter(args[0]) // TODO make this optional with autonaming of column (increment or from first line)

	src, closer, err := common.GetSource(args, 1)
	if err != nil {
		return err
	}
	defer closer()

	sep := " " // TODO add an optional flag to change this default
	splitter := common.SpaceSplitter
	if sep != " " {
		splitter = func(rawValues string) []string {
			return cleannedSplit(rawValues, sep)
		}
	}

	skipLines := 0 // TODO add an optional flag to change this default
	return lineToJson(columns, splitter, skipLines, src)
}

func cleannedSplit(rawValues string, sep string) []string {
	splitted := strings.Split(rawValues, sep)
	for index, value := range splitted {
		splitted[index] = strings.TrimSpace(value)
	}
	return slices.Clip(splitted)
}

func lineToJson(columns []string, splitter func(string) []string, skipLines int, src *os.File) error {
	scanner := bufio.NewScanner(src)
	for skipped := 0; skipped < skipLines; skipped++ {
		scanner.Scan()
	}

	capacity := len(columns)
	encoder := json.NewEncoder(os.Stdout)
	for scanner.Scan() {
		index := 0
		current := make(map[string]string, capacity)
		for _, value := range splitter(scanner.Text()) {
			if value != "" {
				current[columns[index]] = value
				index++
			}
		}
		if err := encoder.Encode(current); err != nil {
			return err
		}
	}
	return scanner.Err()
}
