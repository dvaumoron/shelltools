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
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/dvaumoron/shelltools/common"
	"github.com/spf13/cobra"
)

type attrAndData[T cmp.Ordered] struct {
	attr T
	data []byte
}

type byAttr[T cmp.Ordered] []attrAndData[T]

func (a byAttr[T]) Len() int           { return len(a) }
func (a byAttr[T]) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byAttr[T]) Less(i, j int) bool { return a[i].attr < a[j].attr }

type byAttrDesc[T cmp.Ordered] []attrAndData[T]

func (a byAttrDesc[T]) Len() int           { return len(a) }
func (a byAttrDesc[T]) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byAttrDesc[T]) Less(i, j int) bool { return a[i].attr > a[j].attr }

var extractAsNumber bool
var descOrder bool

func main() {
	cmd := cobra.Command{
		Use:   "jsonorderby COLUMN [FILE]",
		Short: "jsonorderby sort JSON object from FILE on COLUMN field.",
		Long:  "todo",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  jsonOrderByWithInit,
	}

	cmdFlags := cmd.Flags()
	cmdFlags.BoolVarP(&extractAsNumber, "number", "n", false, "process values in ordering column as number")
	cmdFlags.BoolVarP(&descOrder, "desc", "d", false, "sort in descending order")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func jsonOrderByWithInit(cmd *cobra.Command, args []string) error {
	column := args[0]

	src, closer, err := common.GetSource(args, 1)
	if err != nil {
		return err
	}
	defer closer()

	if extractAsNumber {
		return orderBy(column, src, func(jsonObject map[string]any) float64 {
			return extractFloat(jsonObject, column)
		})
	}
	return orderBy(column, src, func(jsonObject map[string]any) string {
		return common.ExtractString(jsonObject, column)
	})
}

func orderBy[T cmp.Ordered](column string, src *os.File, extracter func(map[string]any) T) error {
	var attrAndDatas []attrAndData[T]
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		b := scanner.Bytes()
		var jsonObject map[string]any
		if err := json.Unmarshal(b, &jsonObject); err != nil {
			return err
		}
		attrAndDatas = append(attrAndDatas, attrAndData[T]{attr: extracter(jsonObject), data: b})
	}

	if descOrder {
		sort.Sort(byAttrDesc[T](attrAndDatas))
	} else {
		sort.Sort(byAttr[T](attrAndDatas))
	}

	endLine := []byte{'\n'}
	for _, value := range attrAndDatas {
		if _, err := os.Stdout.Write(value.data); err != nil {
			return err
		}
		if _, err := os.Stdout.Write(endLine); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func extractFloat(jsonObject map[string]any, column string) float64 {
	value := jsonObject[column]
	switch casted := value.(type) {
	case bool:
		if casted {
			return 1
		}
	case float64:
		return casted
	case string:
		parsed, _ := strconv.ParseFloat(casted, 64)
		return parsed
	}
	return 0
}
