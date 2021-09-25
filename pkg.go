// Copyright 2021 The GoRE Authors. All rights reserved.
// Use of this source code is governed by the license that
// can be found in the LICENSE file.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cheynewallace/tabby"
	gore "github.com/goretk/gore"
	"github.com/spf13/cobra"
)

func init() {
	// Flags
	var includeSTD bool
	var includeVendor bool
	var includeUnknown bool
	var includeFilepath bool

	pkgCmd := &cobra.Command{
		Use:     "packages path/to/go/file",
		Aliases: []string{"pkg", "pkgs", "p"},
		Short:   "List packages.",
		Long:    longPkgHelp,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			listPackages(args[0], listPackagesOptions{
				std:  includeSTD,
				vend: includeVendor,
				unk:  includeUnknown,
				fp:   includeFilepath,
			})
		},
	}

	pkgCmd.Flags().BoolVarP(&includeSTD, "std", "s", false, "Include standard library packages.")
	pkgCmd.Flags().BoolVarP(&includeVendor, "vendor", "v", false, "Include 3rd party/vendor packages.")
	pkgCmd.Flags().BoolVarP(&includeUnknown, "unknown", "u", false, "Include unidentified packages.")
	pkgCmd.Flags().BoolVarP(&includeFilepath, "filepath", "f", false, "Include the package's filepath.")

	rootCmd.AddCommand(pkgCmd)
}

type listPackagesOptions struct {
	std  bool
	vend bool
	unk  bool
	fp   bool
}

func listPackages(fileStr string, opts listPackagesOptions) {
	fp, err := filepath.Abs(fileStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse the filepath: %s.\n", err)
		os.Exit(1)
	}

	f, err := gore.Open(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when opening the file: %s.\n", err)
		os.Exit(1)
	}
	defer f.Close()

	packages, err := f.GetPackages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when parsing packages: %s.\n", err)
		os.Exit(1)
	}
	printPackages("Packages", packages, opts.fp)

	if opts.vend {
		// Err check not needed since the parsing was has already been checked.
		vn, _ := f.GetVendors()
		printPackages("Vendors", vn, opts.fp)
	}

	if opts.std {
		std, _ := f.GetSTDLib()
		printPackages("Standard Library Packages", std, opts.fp)
	}

	if opts.unk {
		unk, _ := f.GetUnknown()
		printPackages("Unknown Packages", unk, opts.fp)
	}
}

func printPackages(header string, pkgs []*gore.Package, fp bool) {
	fmt.Printf("%s:\n", header)
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].Name < pkgs[j].Name
	})

	t := tabby.New()
	if fp {
		t.AddHeader("Name", "Version", "Path")
	} else {
		t.AddHeader("Name", "Version")
	}

	for _, p := range pkgs {
		var ver string
		// In Go mod projects, the version is appended at the end of
		// the file path string separated by a "@" character.
		i := strings.LastIndex(p.Filepath, "@")
		if i != -1 && i != len(p.Filepath)+1 {
			ver = strings.Split(p.Filepath[i+1:], "/")[0]
		}

		if fp {
			t.AddLine(p.Name, ver, p.Filepath)
		} else {
			t.AddLine(p.Name, ver)
		}
	}
	t.Print()
	fmt.Printf("\n")
}

const longPkgHelp = `List Packages

The different Go packages used in the binary is extracted. Redress tries by
default to only display the main package and related packages and skips
standard library and 3rd party library packages. Sometimes though, redress
fails to classify a package. In this case, the unclassified packages can be
printed by also provide the unknown flag.

To also include the standard library packages, use the standard library flag.

For 3rd party packages, use the vendor flag.

The folder locations for the packages can also be included by using the
filepath flag.
`
