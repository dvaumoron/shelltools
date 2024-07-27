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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dvaumoron/shelltools/pkg/common"
)

func main() {
	cmd := cobra.Command{
		Use:   "distinctline [FILE]",
		Short: "distinctline echo input without repeated values.",
		Long: `distinctline echo input without repeated values,
without FILE or if FILE is -, read from standard input`,
		Args: cobra.MaximumNArgs(1),
		RunE: distinctLineWithInit,
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func distinctLineWithInit(cmd *cobra.Command, args []string) error {
	src, closer, err := common.GetSource(args, 0)
	if err != nil {
		return err
	}
	defer closer()

	return distinctLine(src)
}

func distinctLine(src *os.File) error {
	endLine := []byte{'\n'}
	values := map[string]struct{}{}
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		value := scanner.Text()
		if _, ok := values[value]; ok {
			continue
		}
		values[value] = struct{}{}

		if _, err := os.Stdout.WriteString(value); err != nil {
			return err
		}
		if _, err := os.Stdout.Write(endLine); err != nil {
			return err
		}

	}
	return scanner.Err()
}
