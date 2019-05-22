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
	"bytes"
	"github.com/spf13/cobra"
	"io"
	"testing"
)

func newVersionCommand(writer io.Writer) *cobra.Command {
	testRootCommand := &cobra.Command{
		Use: "root",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	testRootCommand.AddCommand(newVersionCmd(writer))

	return testRootCommand
}

func TestVersion(t *testing.T) {
	buf := new(bytes.Buffer)

	root := newVersionCommand(buf)
	root.SetOutput(buf)
	root.SetArgs([]string{"version"})

	err := root.Execute()

	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())
	}

	result := buf.String()
	if result != Version+"\n" {
		t.Errorf("Expected: [%s], got: [%s]", Version, result)
	}
}
