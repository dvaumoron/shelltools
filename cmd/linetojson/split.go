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
	"errors"
	"slices"
	"strings"
	"unicode"

	"github.com/dvaumoron/shelltools/pkg/common"
)

var (
	columnEnds []int

	errColumnNotFound = errors.New("column name not found")
)

func lengthSplitter(rawValues string) []string {
	lastColumnIndex := len(columnEnds)
	splitted := make([]string, lastColumnIndex+1)
	start, runeValues := 0, []rune(rawValues)
	for index, end := range columnEnds {
		splitted[index] = strings.TrimSpace(string(runeValues[start:end]))
		start = end
	}
	splitted[lastColumnIndex] = strings.TrimSpace(string(runeValues[start:]))
	return splitted
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

func trimSplitter(rawValues string) []string {
	splitted := strings.Split(rawValues, separator)
	common.TrimSlice(splitted)
	return slices.Clip(splitted)
}

func initColumnEndsFromName(rawValues string, names []string) error {
	columnsNumber := len(names)
	columnEnds = make([]int, 0, columnsNumber)
	for i := 1; i < columnsNumber; i++ {
		name := names[i]
		index := strings.Index(rawValues, name)
		if index == -1 {
			return errColumnNotFound
		}
		runeIndex := len([]rune(rawValues[:index]))
		columnEnds = append(columnEnds, runeIndex)
	}

	return nil
}

func initColumnEndsFromSpace(rawValues string, skippeds []int) {
	prevSpace := false
	for index, c := range []rune(rawValues) {
		isSpace := unicode.IsSpace(c)
		if prevSpace && !isSpace {
			columnEnds = append(columnEnds, index)
		}
		prevSpace = isSpace
	}

	if len(skippeds) != 0 {
		skippedSet := map[int]struct{}{}
		for _, skipped := range skippeds {
			skippedSet[skipped] = struct{}{}
		}

		initialEnds := columnEnds
		columnEnds = make([]int, 0, len(initialEnds))
		for index, end := range initialEnds {
			if _, skipped := skippedSet[index]; !skipped {
				columnEnds = append(columnEnds, end)
			}
		}
	}
}
