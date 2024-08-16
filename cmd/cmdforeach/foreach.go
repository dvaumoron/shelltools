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
	"fmt"
	"os"

	"github.com/dvaumoron/shelltools/pkg/cmdproxy"
	"github.com/dvaumoron/shelltools/pkg/common"
)

const (
	errorMessage = `Error: %[1]s
Usage:
  cmdforeach [CMD] [ARG ...] FILE [flags]

Flags:
  -h, --help   help for cmdforeach

%[1]s
`

	helpMessage = `cmdforeach run one CMD augmented with each line from FILE,
if FILE is -, read from standard input

Usage:
  cmdforeach [CMD] [ARG ...] FILE [flags]

Flags:
  -h, --help   help for cmdforeach`
)

func main() {
	if err := cmdForEachWithInit(os.Args[1:]); err != nil {
		fmt.Printf(errorMessage, err)
		os.Exit(1)
	}
}

func cmdForEachWithInit(args []string) error {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println(helpMessage)

			return nil
		}
	}

	argLen := len(args)
	if argLen < 2 {
		return fmt.Errorf("requires at least 2 arg(s), only received %d", argLen)
	}

	last := argLen - 1
	src, closer, err := common.GetSource(args, last)
	if err != nil {
		return err
	}
	defer closer()

	return cmdForEach(args[0], args[1:last], src)
}

func cmdForEach(cmdName string, cmdArgs []string, src *os.File) error {
	lines, err := common.TrimmedLines(src)
	if err != nil {
		return err
	}

	last := len(cmdArgs)
	cmdArgs = append(cmdArgs, "")
	for _, arg := range lines {
		cmdArgs[last] = arg

		cmdproxy.Run(cmdName, cmdArgs)
	}

	return nil
}
