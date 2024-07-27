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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/spf13/cobra"

	"github.com/dvaumoron/shelltools/pkg/common"
)

var errNoSep = errors.New("no '=' separator")

type Rule struct {
	Name      string
	Transform func(any) (any, error)
}

func makeRule(name string, expression string) (Rule, error) {
	prog, err := expr.Compile(expression)
	if err != nil {
		return Rule{}, err
	}

	return Rule{
		Name: name,
		Transform: func(value any) (any, error) {
			return expr.Run(prog, value)
		},
	}, nil
}

func main() {
	cmd := cobra.Command{
		Use:   "jsontransform [name=EXPRESSION ...] FILE",
		Short: "jsontransform transform JSON object from FILE with EXPRESSION as rules.",
		Long: `jsontransform transform JSON object from FILE with EXPRESSION as rules,
if FILE is -, read from standard input,
to know which EXPRESSION is accepted : see https://expr-lang.org/docs/language-definition`,
		Args: cobra.MinimumNArgs(2),
		RunE: jsonTransformWithInit,
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func jsonTransformWithInit(cmd *cobra.Command, args []string) error {
	argLen := len(args)
	last := argLen - 1

	rules, err := parseRules(args[:last])
	if err != nil {
		return err
	}

	src, closer, err := common.GetSource(args, last)
	if err != nil {
		return err
	}
	defer closer()

	return jsonTransform(rules, src)
}

func parseRules(expressions []string) ([]Rule, error) {
	rules := make([]Rule, 0, len(expressions))
	for _, namedExpression := range expressions {
		i := strings.IndexByte(namedExpression, '=')
		if i == -1 {
			return nil, errNoSep
		}

		rule, err := makeRule(namedExpression[:i], namedExpression[i+1:])
		if err != nil {
			return nil, err
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func jsonTransform(rules []Rule, src *os.File) error {
	scanner := bufio.NewScanner(src)
	encoder := json.NewEncoder(os.Stdout)
	for scanner.Scan() {
		var jsonValue any
		err := json.Unmarshal(scanner.Bytes(), &jsonValue)
		if err != nil {
			return err
		}

		newObject := map[string]any{}
		for _, rule := range rules {
			if newObject[rule.Name], err = rule.Transform(jsonValue); err != nil {
				return err
			}
		}

		if err = encoder.Encode(newObject); err != nil {
			return err
		}
	}
	return scanner.Err()
}
