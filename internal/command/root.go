// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"github.com/ajpantuso/pen-finder/internal/command/start"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "pen-finder [command]",
	}
	cmd.AddCommand(start.NewCommand())

	return cmd
}
