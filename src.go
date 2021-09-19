// Copyright 2019-2021 The GoRE Authors. All rights reserved.
// Use of this source code is governed by the license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goretk/gore"
	"github.com/spf13/cobra"
)

func init() {
	// Flags
	var includeSTD bool
	var includeVendor bool
	var includeUnknown bool
	var includedList []string

	pkgSrc := &cobra.Command{
		Use:     "source path/to/go/file",
		Aliases: []string{"src", "s"},
		Short:   "Source Code Projection.",
		Long:    longSrcHelp,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			listSrc(args[0], listSrcOptions{
				std:     includeSTD,
				vend:    includeVendor,
				unk:     includeUnknown,
				include: includedList,
			})
		},
	}

	pkgSrc.Flags().BoolVarP(&includeSTD, "std", "s", false, "Include standard library packages.")
	pkgSrc.Flags().BoolVarP(&includeVendor, "vendor", "v", false, "Include 3rd party/vendor packages.")
	pkgSrc.Flags().BoolVarP(&includeUnknown, "unknown", "u", false, "Include unidentified packages.")
	pkgSrc.Flags().StringSliceVarP(&includedList, "include", "i", nil, "Include the following packages. Can be provided as a comma-separated list or via providing the flag multiple times.")

	rootCmd.AddCommand(pkgSrc)
}

type listSrcOptions struct {
	std     bool
	vend    bool
	unk     bool
	include []string
}

func listSrc(fileStr string, opts listSrcOptions) {
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

	// Err check not needed since the parsing was has already been checked.
	vn, _ := f.GetVendors()
	if opts.vend {
		packages = append(packages, vn...)
	} else {
		for _, i := range opts.include {
			for _, v := range vn {
				if v.Name == i {
					packages = append(packages, v)
				}
			}
		}
	}

	std, _ := f.GetSTDLib()
	if opts.std {
		packages = append(packages, std...)
	} else {
		for _, i := range opts.include {
			for _, v := range std {
				if v.Name == i {
					packages = append(packages, v)
				}
			}
		}
	}

	unk, _ := f.GetUnknown()
	if opts.unk {
		packages = append(packages, unk...)
	} else {
		for _, i := range opts.include {
			for _, v := range unk {
				if v.Name == i {
					packages = append(packages, v)
				}
			}
		}
	}

	printFolderStructures(packages)
}

func printFolderStructures(pkgs []*gore.Package) {
	for i, p := range pkgs {
		if i != 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("Package %s: %s\n", p.Name, p.Filepath)
		for _, sf := range p.GetSourceFiles() {
			sf.Postfix = "\t"
			fmt.Printf("%s\n", sf)
		}
	}
}

const longSrcHelp = `Source Code Projection

Construct a source code tree layout based on the metadata found in the binary.
The output includes the package name and its folder location at compile time.
For each file, the functions defined within are printed. The output also
includes auto generated functions produced by the compiler. For each function,
redress tries to guess the starting and ending line number.
	Folder -> File -> Function -> Line

By default, standard library and 3rd party packages are excluded but can be
included by providing the flags "std", "vendor", and/or "unknown". It is also
possible to include individual packages with the "include" flag.
`
