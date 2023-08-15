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

	"github.com/antonmedv/expr"
	"github.com/dvaumoron/shelltools/common"
	"github.com/spf13/cobra"
)

func main() {
	cmd := cobra.Command{
		Use:   "jsonwhere EXPRESSION [FILE]",
		Short: "jsonwhere filter JSON object from FILE with EXPRESSION as predicate.",
		Long: `jsonwhere filter JSON object from FILE with EXPRESSION as predicate,
without FILE or if FILE is -, read from standard input,
to know which EXPRESSION is accepted : see https://expr.medv.io/docs/Language-Definition`,
		Args: cobra.RangeArgs(1, 2),
		RunE: jsonWhereWithInit,
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func jsonWhereWithInit(cmd *cobra.Command, args []string) error {
	pred, err := parsePredicate(args[0])
	if err != nil {
		return err
	}

	src, closer, err := common.GetSource(args, 1)
	if err != nil {
		return err
	}
	defer closer()

	return jsonWhere(pred, src)
}

func parsePredicate(expression string) (func(any) bool, error) {
	prog, err := expr.Compile(expression)
	if err != nil {
		return nil, err
	}

	return func(value any) bool {
		output, _ := expr.Run(prog, value)
		casted, _ := output.(bool)
		return casted
	}, nil
}

func jsonWhere(pred func(any) bool, src *os.File) error {
	endLine := []byte{'\n'}
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		b := scanner.Bytes()
		var jsonValue any
		if err := json.Unmarshal(b, &jsonValue); err != nil {
			return err
		}
		if pred(jsonValue) {
			if _, err := os.Stdout.Write(b); err != nil {
				return err
			}
			if _, err := os.Stdout.Write(endLine); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}
