// Copyright 2021 The GoRE Authors. All rights reserved.
// Use of this source code is governed by the license that
// can be found in the LICENSE file.
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	gore "github.com/goretk/gore"
	"github.com/spf13/cobra"
)

func init() {
	// Flags
	var includeStd bool
	var includeVendor bool
	var includeMethods bool
	var forceGoVersion string

	typCmd := &cobra.Command{
		Use:       "types {struct|interface|all} path/to/go/file",
		Aliases:   []string{"type", "typ", "t"},
		Short:     "List types.",
		Long:      longTypesHelp,
		Args:      cobra.ExactArgs(2),
		ValidArgs: []string{"struct", "interface", "all"},
		Run: func(cmd *cobra.Command, args []string) {
			// Find what types should be printed
			opt := listTypesOptions{}
			fpArg := 0
			for i, a := range args {
				switch a {
				case "struct":
					opt.structs = true
				case "interface":
					opt.interfaces = true
				case "all":
					opt.all = true
				default:
					fpArg = i
				}
			}

			// Check mode.
			num := 0
			if opt.all {
				num++
			}
			if opt.interfaces {
				num++
			}
			if opt.structs {
				num++
			}

			if num > 1 {
				fmt.Fprintf(os.Stderr, "struct, interface and all are mutually exclusive. Only one can be used at the time.\n")
				os.Exit(1)
			} else if num == 0 {
				fmt.Fprintf(os.Stderr, "One of: struct, interface or all needs to be used.\n")
				os.Exit(1)
			}

			opt.std = includeStd
			opt.vendor = includeVendor
			opt.methods = includeMethods
			opt.goversion = forceGoVersion

			listTypes(args[fpArg], opt)
		},
	}

	typCmd.Flags().BoolVarP(&includeStd, "std", "s", false, "Include standard library packages.")
	typCmd.Flags().BoolVarP(&includeVendor, "vendor", "v", false, "Include 3rd party/vendor packages.")
	typCmd.Flags().BoolVarP(&includeMethods, "methods", "m", false, "Include method definitions.")
	typCmd.Flags().StringVar(&forceGoVersion, "version", assumedGoVersion, "Fallback compiler version.")

	typeOffsetCmd := &cobra.Command{
		Use:   "offset address path/to/file",
		Short: "Print type at the given address.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			addr, err := strconv.ParseUint(args[0], 0, strconv.IntSize)
			if err != nil {
				fmt.Printf("Bad address format: %s.\n", err)
				return
			}
			file, err := gore.Open(args[1])
			if err != nil {
				fmt.Println("Error when opening the file:", err)
				return
			}
			defer file.Close()

			// If the user has provided a specific compiler version that we should assume,
			// we force it.
			if forceGoVersion != "" {
				err = file.SetGoVersion(forceGoVersion)
				if err != nil {
					fmt.Println("Error when setting the assumed Go version:", err)
					return
				}
			}

			lookupType(file, addr)
		},
	}
	typeOffsetCmd.Flags().StringVar(&forceGoVersion, "version", "", "Fallback compiler version.")

	typCmd.AddCommand(typeOffsetCmd)
	rootCmd.AddCommand(typCmd)
}

type listTypesOptions struct {
	structs    bool
	interfaces bool
	all        bool
	std        bool
	vendor     bool
	methods    bool
	goversion  string
}

func listTypes(fileStr string, opts listTypesOptions) {
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

	typs, err := f.GetTypes()
	if err == gore.ErrNoGoVersionFound {
		// Force the assumed version and try again.
		f.SetGoVersion(opts.goversion)
		typs, err = f.GetTypes()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when enumerating types: %s.\n", err)
		os.Exit(1)
	}

	printTypes(typs, opts)
}

func printTypes(typs []*gore.GoType, opts listTypesOptions) {
	if opts.all {
		opts.structs = true
		opts.interfaces = true
	}

	buf := &bytes.Buffer{}

	for _, typ := range typs {
		// Try to filter out vendor packages if needed.
		if !opts.vendor && strings.Contains(typ.PackagePath, "/vendor/") {
			continue
		}

		// Try to filter out std packages if needed.
		if !opts.std &&
			(gore.IsStandardLibrary(typ.PackagePath) ||
				strings.HasPrefix(typ.Name, "map.") ||
				strings.HasPrefix(typ.Name, "*map.")) {
			continue
		}

		if (opts.structs) && (typ.Kind == reflect.Struct) {
			s := gore.StructDef(typ)
			out, err := format.Source([]byte(strings.ReplaceAll(s, ".", "____")))
			if err != nil {
				fmt.Fprintf(buf, "%s\n", s)
			} else {
				fmt.Fprintf(buf, "%s\n", strings.ReplaceAll(string(out), "____", "."))
			}
			if opts.methods && len(typ.Methods) > 0 {
				fmt.Fprintf(buf, "%s\n\n", gore.MethodDef(typ))
			} else {
				fmt.Fprintf(buf, "\n")
			}
			continue
		}
		if (opts.interfaces) && (typ.Kind == reflect.Interface) {
			i := gore.InterfaceDef(typ)
			out, err := format.Source([]byte(strings.ReplaceAll(i, ".", "____")))
			if err != nil {
				fmt.Fprintf(buf, "%s\n", i)
			} else {
				fmt.Fprintf(buf, "%s\n", strings.ReplaceAll(string(out), "____", "."))
			}
			continue
		}
		if opts.all {
			fmt.Fprintf(buf, "%s\n", typ)
			if opts.methods && len(typ.Methods) > 0 {
				fmt.Fprintf(buf, "%s\n\n", gore.MethodDef(typ))
			} else {
				fmt.Fprintf(buf, "\n")
			}
		}
	}

	fmt.Println(buf.String())
}

const longTypesHelp = `List Types

Redress can display different type data found in the binary. Interfaces can be
extracted with the "interface" argument while structures can be extracted with
the "struct" argument.

By default, standard library types are filtered out. These can be included
by also providing the standard library flag.

Method definitions for types can be included by using the method flag.

It is also possible to print all types in the binary by using the "all"
argument.

Redress tries to detect the version of the compiler that produced the binary.
If this process fails, a fallback version can be provided.
`
