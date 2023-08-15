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

	"github.com/dvaumoron/shelltools/common"
	"github.com/spf13/cobra"
)

type columnNamer interface {
	Init([]string) (int, bool)
	Name(int) string
}

type numberNamer struct {
	names []string
}

func (n *numberNamer) Init(values []string) (int, bool) {
	size := len(values)
	n.Name(size - 1)
	return size, true
}

func (n *numberNamer) Name(index int) string {
	if size := len(n.names); size <= index {
		for targetSize := index + 1; size < targetSize; size++ {
			n.names = append(n.names, "col"+strconv.Itoa(size))
		}
	}
	return n.names[index]
}

type fromFirstNamer struct {
	numberNamer // same implementation of Name
}

func (f *fromFirstNamer) Init(values []string) (int, bool) {
	size := len(values)
	f.names = values
	return size, false
}

var separator string
var columns []string
var fromFirst bool

func main() {
	cmd := cobra.Command{
		Use:   "linetojson [FILE]",
		Short: "linetojson convert each line from FILE in a JSON object.",
		Long: `linetojson convert each line from FILE in a JSON object, without flag:
- create the column name as 'col#'
- use space as separator`,
		Args: cobra.MaximumNArgs(1),
		RunE: lineToJsonWithInit,
	}

	cmdFlags := cmd.Flags()
	cmdFlags.StringVarP(&separator, "separator", "s", " ", "separator for value in line")
	cmdFlags.BoolVarP(&fromFirst, "first", "f", false, "initialize column name with first line")
	cmdFlags.StringSliceVarP(&columns, "columns", "c", nil, "name of the columns (comma separated)")
	cmd.MarkFlagsMutuallyExclusive("first", "columns")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func lineToJsonWithInit(cmd *cobra.Command, args []string) error {
	trimSlice(columns)

	src, closer, err := common.GetSource(args, 0)
	if err != nil {
		return err
	}
	defer closer()

	splitter := common.SpaceSplitter
	if separator != " " {
		splitter = trimSplitter
	}

	var namer columnNamer = &numberNamer{names: columns} // if not enough name, fall back to 'col#'
	if fromFirst {
		namer = &fromFirstNamer{}
	}
	return lineToJson(namer, splitter, src)
}

func trimSplitter(rawValues string) []string {
	splitted := strings.Split(rawValues, separator)
	trimSlice(splitted)
	return slices.Clip(splitted)
}

func trimSlice(values []string) {
	for index, value := range values {
		values[index] = strings.TrimSpace(value)
	}
}

func lineToJson(namer columnNamer, splitter func(string) []string, src *os.File) error {
	scanner := bufio.NewScanner(src)
	encoder := json.NewEncoder(os.Stdout)
	if scanner.Scan() {
		splitted := splitter(scanner.Text())
		capacity, add := namer.Init(splitted)
		if add {
			first := toJsonObject(splitted, capacity, namer)
			if err := encoder.Encode(first); err != nil {
				return err
			}
		}

		for scanner.Scan() {
			splitted = splitter(scanner.Text())
			current := toJsonObject(splitted, capacity, namer)
			if err := encoder.Encode(current); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

func toJsonObject(splitted []string, capacity int, namer columnNamer) map[string]string {
	jsonObject := make(map[string]string, capacity)
	for index, value := range splitted {
		if value != "" {
			jsonObject[namer.Name(index)] = value
		}
	}
	return jsonObject
}
