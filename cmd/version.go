/*
 * Nuts registry
 * Copyright (C) 2019 Nuts community
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

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Creates a new Version command printing to Stdout
func NewVersionCmd() *cobra.Command {
	return newVersionCmd(os.Stdout)
}

// Creates a new Version command printing to the given writer
func newVersionCmd(writer io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of the Nuts registry",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(writer, VERSION)
		},
	}
}
