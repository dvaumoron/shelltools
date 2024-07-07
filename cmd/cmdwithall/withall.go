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
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dvaumoron/shelltools/pkg/common"
)

func main() {
	cmd := cobra.Command{
		Use:   "cmdwithall [CMD] [ARG ...] [FILE]",
		Short: "cmdwithall run CMD adding more args from FILE.",
		Long: `cmdwithall run CMD adding more args from FILE,
if FILE is -, read from standard input`,
		Args: cobra.MinimumNArgs(2),
		RunE: cmdWithAllWithInit,
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func cmdWithAllWithInit(cmd *cobra.Command, args []string) error {
	last := len(args) - 1
	src, closer, err := common.GetSource(args, last)
	if err != nil {
		return err
	}
	defer closer()

	return cmdWithAll(args[0], args[1:last], src)
}

func cmdWithAll(cmdName string, cmdArgs []string, src *os.File) error {
	lines, err := trimmedLines(src)
	if err != nil {
		return err
	}

	cmdArgs = append(cmdArgs, lines...)

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Println("Failure during", cmdName, "call :", err)
	}
	return nil
}

func trimmedLines(src *os.File) ([]string, error) {
	splitted := []string{}
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		splitted = append(splitted, strings.TrimSpace(scanner.Text()))
	}
	return splitted, scanner.Err()
}
