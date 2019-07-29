// Copyright 2019 The GoRE.tk Authors. All rights reserved.
// Use of this source code is governed by the license that
// can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"

	"github.com/TcM1911/r2g2"
)

const (
	assumedGoVersion = "go1.12"
)

type option struct {
	printPackages        *bool
	printStdLibPackages  *bool
	printVendorPackages  *bool
	printUnknownPackages *bool
	includeFilepath      *bool
	printSourceTree      *bool
	printTypes           *bool
	printStructs         *bool
	printMethods         *bool
	printIntefaces       *bool
	lookupType           *int
	printCompiler        *bool
	printVersion         *bool
}

// This is set at compile time.
var redressVersion string

var options option

func init() {
	if r2g2.CheckForR2Pipe() {
		options.lookupType = flag.Int("type", 0, "Lookup the Go definition for a type")
	} else {
		options.printPackages = flag.Bool("pkg", false, "List packages")
		options.printStdLibPackages = flag.Bool("std", false, "Include standard library packages")
		options.printVendorPackages = flag.Bool("vendor", false, "Include vendor packages")
		options.printUnknownPackages = flag.Bool("unknown", false, "Include unknown packages")
		options.includeFilepath = flag.Bool("filepath", false, "Include file path for packages")
		options.printSourceTree = flag.Bool("src", false, "Print source tree")
		options.printTypes = flag.Bool("type", false, "Print all type information")
		options.printStructs = flag.Bool("struct", false, "Print structs")
		options.printIntefaces = flag.Bool("interface", false, "Print interfaces")
		options.printCompiler = flag.Bool("compiler", false, "Print information")
	}
	options.printMethods = flag.Bool("method", false, "Print type's methods")
	options.printVersion = flag.Bool("version", false, "Print redress version")
	flag.Parse()
}

func main() {
	if *options.printVersion {
		fmt.Printf("Redress version: %s\n", redressVersion)
		return
	}
	if r2g2.CheckForR2Pipe() {
		r2Exec()
	} else {
		standalone()
	}
}
