// Copyright 2019 The GoRE.tk Authors. All rights reserved.
// Use of this source code is governed by the license that
// can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/goretk/gore"
)

func standalone() {
	if len(flag.Args()) != 1 {
		return
	}

	fileStr, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalln("Failed to parse the filepath:", err)
	}

	f, err := gore.Open(fileStr)
	if err != nil {
		log.Fatalln("Error when opening the file:", err)
	}
	defer f.Close()

	// Setting forced version if given
	if *options.forceVersion != "" {
		if err = f.SetGoVersion(*options.forceVersion); err != nil {
			fmt.Println("Failed to set the given Go version:", err)
			return
		}
	}

	if *options.printCompiler {
		cmp, err := f.GetCompilerVersion()
		if err != nil {
			fmt.Println("Error when extracting compiler information:", err)
		} else {
			fmt.Printf("Compiler version: %s (%s)\n", cmp.Name, cmp.Timestamp)
		}
	}

	pkgs, err := f.GetPackages()
	if err != nil {
		fmt.Println(err)
		return
	}

	if *options.printTypes || *options.printStructs || *options.printIntefaces {
		typs, err := f.GetTypes()
		if err == gore.ErrNoGoVersionFound {
			// Force the assumed version and try again.
			f.SetGoVersion(assumedGoVersion)
			typs, err = f.GetTypes()
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		printTypes(typs, &options)
	}

	if *options.printPackages {
		pkgs, err := f.GetPackages()
		if err != nil {
			fmt.Println(err)
			return
		}
		printPackages("Packages", pkgs)
		if *options.printVendorPackages {
			vendorpkgs, err := f.GetVendors()
			if err != nil {
				fmt.Println(err)
				return
			}
			printPackages("Vendors", vendorpkgs)
		}
		if *options.printStdLibPackages {
			stdpkgs, err := f.GetSTDLib()
			if err != nil {
				fmt.Println(err)
				return
			}
			printPackages("Standard Libraries", stdpkgs)
		}
		if *options.printUnknownPackages {
			unknownpkgs, err := f.GetUnknown()
			if err != nil {
				fmt.Println(err)
				return
			}
			printPackages("Unknown Libraries", unknownpkgs)
		}
	}

	if *options.printUnknownPackages {
		upkg, _ := f.GetUnknown()
		pkgs = append(pkgs, upkg...)
	}

	if *options.printVendorPackages {
		vpkg, _ := f.GetVendors()
		pkgs = append(pkgs, vpkg...)
	}

	if *options.printStdLibPackages {
		std, _ := f.GetSTDLib()
		pkgs = append(pkgs, std...)
	}

	if *options.printSourceTree {
		printFolderStructures(pkgs)
	}
}

func printPackages(header string, pkgs []*gore.Package) {
	fmt.Printf("%s:\n", header)
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].Name < pkgs[j].Name
	})
	for _, p := range pkgs {
		if *options.includeFilepath {
			fmt.Printf("%s | %s\n", p.Name, p.Filepath)
		} else {
			fmt.Printf("%s\n", p.Name)
		}
	}
	fmt.Println("")
}

func printFolderStructures(pkgs []*gore.Package) {
	for _, p := range pkgs {
		fmt.Printf("Package %s: %s\n", p.Name, p.Filepath)
		for _, sf := range p.GetSourceFiles() {
			sf.Postfix = "\t"
			fmt.Printf("%s\n", sf)
		}
	}
}

func printTypes(typs []*gore.GoType, opts *option) {
	if *opts.printTypes {
		*opts.printStructs = true
		*opts.printIntefaces = true
	}
	for _, typ := range typs {
		// Try to filter out vendor packages if needed.
		if !*opts.printVendorPackages && strings.Contains(typ.PackagePath, "/vendor/") {
			continue
		}

		// Try to filter out std packages if needed.
		if !*opts.printStdLibPackages &&
			(gore.IsStandardLibrary(typ.PackagePath) ||
				strings.HasPrefix(typ.Name, "map.") ||
				strings.HasPrefix(typ.Name, "*map.")) {
			continue
		}

		if (*opts.printStructs) && (typ.Kind == reflect.Struct) {
			fmt.Println(gore.StructDef(typ))
			if *options.printMethods && len(typ.Methods) > 0 {
				fmt.Println(gore.MethodDef(typ) + "\n")
			} else {
				fmt.Println("")
			}
			continue
		}
		if (*opts.printIntefaces) && (typ.Kind == reflect.Interface) {
			fmt.Println(gore.InterfaceDef(typ) + "\n")
			continue
		}
		if *opts.printTypes {
			fmt.Println(typ)
			if *options.printMethods && len(typ.Methods) > 0 {
				fmt.Println(gore.MethodDef(typ) + "\n")
			} else {
				fmt.Println("")
			}
		}
	}
}
