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
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/tofuutils/tenv/v2/pkg/reversecmp"

	"github.com/dvaumoron/shelltools/pkg/common"
)

type attrAndData[T any] struct {
	attr T
	data []byte
}

func cmpAttr[T any](cmpFunc func(T, T) int) func(a attrAndData[T], b attrAndData[T]) int {
	return func(a attrAndData[T], b attrAndData[T]) int {
		return cmpFunc(a.attr, b.attr)
	}
}

var (
	descOrder        bool
	extractAsNumber  bool
	extractAsVersion bool
	ignoreCase       bool
	stable           bool
)

func main() {
	cmd := cobra.Command{
		Use:   "jsonorderby COLUMN [FILE]",
		Short: "jsonorderby sort JSON object from FILE on COLUMN field.",
		Long: `jsonorderby sort JSON object from FILE on COLUMN field,
without FILE or if FILE is -, read from standard input`,
		Args: cobra.RangeArgs(1, 2),
		RunE: jsonOrderByWithInit,
	}

	cmdFlags := cmd.Flags()
	cmdFlags.BoolVarP(&extractAsNumber, "number", "n", false, "process values in ordering column as number")
	cmdFlags.BoolVarP(&extractAsVersion, "semver", "v", false, "process values in ordering column as semantic version")
	cmdFlags.BoolVarP(&descOrder, "desc", "d", false, "sort in descending order")
	cmdFlags.BoolVarP(&stable, "stable", "s", false, "use a stable sort")
	cmdFlags.BoolVarP(&ignoreCase, "ignore-case", "i", false, "ignore case in ordering column")
	cmd.MarkFlagsMutuallyExclusive("number", "ignore-case")

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

	switch {
	case extractAsNumber:
		return orderBy(src, func(jsonObject map[string]any) float64 {
			return extractFloat(jsonObject, column)
		}, cmp.Compare[float64])
	case extractAsVersion:
		return orderBy(src,
			func(jsonObject map[string]any) *version.Version {
				return extractVersion(jsonObject, column)
			},
			func(v1 *version.Version, v2 *version.Version) int {
				return v1.Compare(v2)
			})
	case ignoreCase:
		return orderBy(src, func(jsonObject map[string]any) string {
			return strings.ToLower(common.ExtractString(jsonObject, column))
		}, cmp.Compare[string])
	}
	return orderBy(src, func(jsonObject map[string]any) string {
		return common.ExtractString(jsonObject, column)
	}, cmp.Compare[string])
}

func orderBy[T any](src *os.File, extracter func(map[string]any) T, cmpFunc func(T, T) int) error {
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

	err := scanner.Err()
	if err != nil {
		return err
	}

	sortFunc := slices.SortFunc[[]attrAndData[T], attrAndData[T]]
	if stable {
		sortFunc = slices.SortStableFunc[[]attrAndData[T], attrAndData[T]]
	}

	cmpAttrFunc := reversecmp.Reverser(cmpAttr(cmpFunc), descOrder)
	sortFunc(attrAndDatas, cmpAttrFunc)

	endLine := []byte{'\n'}
	for _, value := range attrAndDatas {
		if _, err = os.Stdout.Write(value.data); err != nil {
			return err
		}
		if _, err = os.Stdout.Write(endLine); err != nil {
			return err
		}
	}
	return nil
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

func extractVersion(jsonObject map[string]any, column string) *version.Version {
	v, _ := version.NewVersion(common.ExtractString(jsonObject, column))
	return v
}
