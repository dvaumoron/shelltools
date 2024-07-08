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
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ExtractString(jsonObject map[string]any, column string) string {
	return fmt.Sprint(jsonObject[column])
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

func TrimmedLines(src *os.File) ([]string, error) {
	splitted := []string{}
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		splitted = append(splitted, strings.TrimSpace(scanner.Text()))
	}
	return splitted, scanner.Err()
}

func TrimSlice(values []string) {
	for index, value := range values {
		values[index] = strings.TrimSpace(value)
	}
}
