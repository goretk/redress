// Copyright 2021 The GoRE Authors. All rights reserved.
// Use of this source code is governed by the license that
// can be found in the LICENSE file.

package main

import (
	"github.com/spf13/cobra"
)

var expCmd = &cobra.Command{
	Use:     "experiment",
	Aliases: []string{"exp", "x"},
	Short:   "Experimental functionality",
	Long:    expHelp,
	Hidden:  true,
}

func init() {
	// Register the command.
	rootCmd.AddCommand(expCmd)
}

const expHelp = `Experimental functionality

The following commands are experimental and may change or removed
in the future.
`
