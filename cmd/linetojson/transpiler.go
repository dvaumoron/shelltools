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

	"github.com/spf13/cobra"

	"github.com/dvaumoron/shelltools/pkg/common"
)

type columnNamer interface {
	Init([]string)
	Name(int) string
}

type numberNamer struct {
	names []string
}

func (n *numberNamer) Init(values []string) {
	n.Name(len(values) - 1)
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

func (f *fromFirstNamer) Init(values []string) {
	f.names = values
}

var (
	columns   []string
	fromFirst bool
	skipped   []int
	separator string
	tableMode bool
)

func main() {
	cmd := cobra.Command{
		Use:   "linetojson [FILE]",
		Short: "linetojson convert each line from FILE in a JSON object.",
		Long: `linetojson convert each line from FILE in a JSON object,
without FILE or if FILE is -, read from standard input,
default behaviour :
- create the column name as 'col#'
- use space as separator`,
		Args: cobra.MaximumNArgs(1),
		RunE: lineToJsonWithInit,
	}

	cmdFlags := cmd.Flags()
	cmdFlags.StringVarP(&separator, "separator", "s", " ", "separator for value in line")
	cmdFlags.BoolVarP(&fromFirst, "first", "f", false, "initialize column name with first line")
	cmdFlags.StringSliceVarP(&columns, "columns", "c", nil, "name of the columns (comma separated)")
	cmdFlags.BoolVarP(&tableMode, "table", "t", false, "split fixed size columns")
	cmdFlags.IntSliceVarP(&skipped, "merge", "m", nil, "merge some columns (with following by number (zero based))")
	cmd.MarkFlagsMutuallyExclusive("first", "columns")
	cmd.MarkFlagsMutuallyExclusive("separator", "table")
	cmd.MarkFlagsMutuallyExclusive("columns", "merge")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func lineToJsonWithInit(cmd *cobra.Command, args []string) error {
	common.TrimSlice(columns)

	src, closer, err := common.GetSource(args, 0)
	if err != nil {
		return err
	}
	defer closer()

	splitter := spaceSplitter
	switch {
	case tableMode:
		splitter = lengthSplitter
	case separator != " ":
		splitter = trimSplitter
	}

	var namer columnNamer = &numberNamer{names: columns} // if not enough name, fall back to 'col#'
	if fromFirst {
		namer = &fromFirstNamer{}
	}
	return lineToJson(namer, splitter, src)
}

func lineToJson(namer columnNamer, splitter func(string) []string, src *os.File) error {
	scanner := bufio.NewScanner(src)
	encoder := json.NewEncoder(os.Stdout)
	if scanner.Scan() {
		rawValues := scanner.Text()
		if tableMode {
			if fromFirst || len(columns) == 0 {
				initColumnEndsFromSpace(rawValues, skipped)
			} else if err := initColumnEndsFromName(rawValues, columns); err != nil {
				return err
			}
		}
		splitted := splitter(rawValues)
		namer.Init(splitted)
		if !fromFirst {
			first := toJsonObject(splitted, namer)
			if err := encoder.Encode(first); err != nil {
				return err
			}
		}

		for scanner.Scan() {
			splitted = splitter(scanner.Text())
			current := toJsonObject(splitted, namer)
			if err := encoder.Encode(current); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

func toJsonObject(splitted []string, namer columnNamer) map[string]string {
	jsonObject := make(map[string]string, len(splitted))
	for index, value := range splitted {
		if value != "" {
			jsonObject[namer.Name(index)] = value
		}
	}
	return jsonObject
}
