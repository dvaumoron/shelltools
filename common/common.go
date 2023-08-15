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

package common

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

func TrimSlice(values []string) {
	for index, value := range values {
		values[index] = strings.TrimSpace(value)
	}
}

func GetSource(args []string, pos int) (*os.File, func() error, error) {
	src := os.Stdin
	closer := noActionCloser
	if len(args) > pos {
		if filePath := args[pos]; filePath != "-" {
			var err error
			src, err = os.Open(filePath)
			if err != nil {
				return nil, nil, err
			}
			closer = src.Close
		}
	}
	return src, closer, nil
}

func noActionCloser() error {
	return nil
}

func SpaceSplitter(rawValues string) []string {
	splitted := strings.Split(rawValues, " ")
	values := make([]string, 0, len(splitted))
	for _, value := range splitted {
		if value != "" {
			values = append(values, value)
		}
	}
	return slices.Clip(values)
}

func ExtractString(jsonObject map[string]any, column string) string {
	return fmt.Sprint(jsonObject[column])
}
