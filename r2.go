// Copyright 2019 The GoRE.tk Authors. All rights reserved.
// Use of this source code is governed by the license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/TcM1911/r2g2"
	gore "github.com/goretk/gore"
)

func cleanupName(old string) string {
	newString := strings.Replace(old, " ", "_", -1)
	newString = strings.Replace(newString, "-", "_", -1)
	newString = strings.Replace(newString, ";", "_", -1)
	return newString
}

func r2Exec() {
	// Ensure locations is taken from the pipe.
	r2, err := r2g2.OpenPipe()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	of, err := r2.GetActiveOpenFile()
	if err != nil {
		fmt.Println("Error when getting current open file:", err)
		return
	}

	file, err := gore.Open(of.Path)
	if err != nil {
		fmt.Println("Error when opening the file:", err)
		return
	}
	defer file.Close()

	// Lookup type?
	if *options.lookupType != 0 {
		lookupType(file, uint64(*options.lookupType))
		return
	}

	// Get compiler version
	v, err := file.GetCompilerVersion()
	if err == gore.ErrNoGoVersionFound {
		fmt.Println("Failed to determine the compiler version, assuming", assumedGoVersion)
		file.SetGoVersion(assumedGoVersion)
	} else if err != nil {
		fmt.Println("Failed to determine the compiler version:", err)
	} else {
		fmt.Printf("Compiler version: %s (%s)\n", v.Name, v.Timestamp)
	}

	pkgs, err := file.GetPackages()
	if err != nil {
		fmt.Println("Failed to get packages:", err)
		return
	}

	// Vendors, stdlib and unknown have now been populated so we can ignore the err check.
	vendors, _ := file.GetVendors()
	std, _ := file.GetSTDLib()
	unknown, _ := file.GetUnknown()

	pkgs = append(pkgs, vendors...)
	pkgs = append(pkgs, std...)
	pkgs = append(pkgs, unknown...)

	fmt.Printf("%d packages found.\n", len(pkgs))
	applyFuncSymbols(pkgs, r2)

	// Analyze init and main
	r2.Run("afr @ fcn.main.init")
	r2.Run("afr @ fcn.main.main")

	types, err := file.GetTypes()
	if err != nil {
		fmt.Println("Error when getting types:", err)
		return
	}
	count := 0
	for _, typ := range types {
		if typ.Addr == 0 {
			continue
		}
		r2.NewFlag("sym.type."+cleanupName(typ.Name), typ.Addr)
		count++
	}
	fmt.Printf("%d type symbols found\n", count)
}

func applyFuncSymbols(pkgs []*gore.Package, r2 *r2g2.Client) {
	count := 0
	for _, p := range pkgs {
		for _, f := range p.Functions {
			if f.Offset == uint64(0) {
				continue
			}
			r2.NewFlagWithLength(
				"fcn."+cleanupName(p.Name)+"."+cleanupName(f.Name),
				f.Offset,
				f.End-f.Offset)
			count++
		}
		for _, m := range p.Methods {
			if m.Offset == uint64(0) {
				continue
			}
			r2.NewFlagWithLength(
				"fcn."+cleanupName(p.Name)+"#"+cleanupName(m.Receiver)+"."+cleanupName(m.Name),
				m.Offset,
				m.End-m.Offset)
			count++
		}
	}
	fmt.Printf("%d function symbols found\n", count)
}

func lookupType(f *gore.GoFile, addr uint64) {
	typs, err := f.GetTypes()
	if err != nil {
		fmt.Println("Error when looking up the type:", err)
		return
	}
	for _, typ := range typs {
		if typ.Addr != addr {
			continue
		}
		switch typ.Kind {
		case reflect.Interface:
			fmt.Println(gore.InterfaceDef(typ))
		case reflect.Struct:
			fmt.Println(gore.StructDef(typ))
		default:
			fmt.Println(typ.String())
		}
		if *options.printMethods && len(typ.Methods) != 0 {
			fmt.Println(gore.MethodDef(typ))
		}
	}
}
