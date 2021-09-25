// Copyright 2019-2021 The GoRE Authors. All rights reserved.
// Use of this source code is governed by the license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/cheynewallace/tabby"
	"github.com/spf13/cobra"
)

const (
	assumedGoVersion = "go1.16"
)

// These are set at compile time.
var redressVersion string
var goreVersion string
var compilerVersion string

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "redress",
	Short: banner,
}

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Display redress version information.",
		Run: func(cmd *cobra.Command, args []string) {
			t := tabby.New()
			t.AddLine("Version:", redressVersion)
			t.AddLine("GoRE:", goreVersion)
			t.AddLine("Go:", compilerVersion)
			fmt.Println(banner)
			t.Print()
		},
	})
}

const banner = `______         _                  
| ___ \       | |                 
| |_/ /___  __| |_ __ ___ ___ ___ 
|    // _ \/ _  | '__/ _ / __/ __|
| |\ |  __| (_| | | |  __\__ \__ \
\_| \_\___|\__,_|_|  \___|___|___/
                                 
`
