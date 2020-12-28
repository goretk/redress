// Copyright 2019 The GoRE.tk Authors. All rights reserved.
// Use of this source code is governed by the license that
// can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"reflect"
	"strconv"
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

	// Add source code line info
	if *options.srcLine {
		srcLineInfo(r2, file)
		return
	}

	// Lookup type?
	if *options.lookupType != 0 {
		lookupType(file, uint64(*options.lookupType))
		return
	}

	// Print a string slice
	if *options.resolveStrSlice {
		args := flag.Args()
		if len(args) != 2 {
			fmt.Println("2 arguments are required. Address and slice length")
			return
		}
		address, err := strconv.ParseUint(args[0], 0, 32)
		if err != nil {
			fmt.Println("Failed to parse address argument:", err)
			return
		}
		length, err := strconv.ParseUint(args[1], 0, 32)
		if err != nil {
			fmt.Println("Failed to parse length argument:", err)
			return
		}
		printStringSlice(file, address, length)
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
	fmt.Println("Analyzing all init functions.")
	r2.Run("afr @@ fcn.main~init")

	fmt.Println("Analyzing all main.main.")
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

func printStringSlice(f *gore.GoFile, offset uint64, length uint64) {
	// Number of bytes needed for each string is one word for pointer to data and
	// one word for the length of the data. This needs to be multiplied by the length
	// of the array.
	arrayData, err := f.Bytes(offset, 2*length*uint64(f.FileInfo.WordSize))
	if err != nil {
		fmt.Println("Failed to get string slice:", err)
		return
	}
	r := bytes.NewReader(arrayData)
	for i := 0; i < int(length); i++ {
		strPtr, err := readUintToUint64(r, f.FileInfo)
		if err != nil {
			fmt.Println("Error when reading pointer to string data:", err)
			return
		}
		strLen, err := readUintToUint64(r, f.FileInfo)
		if err != nil {
			fmt.Println("Error when reading string data length:", err)
			return
		}
		strData, err := f.Bytes(strPtr, strLen)
		if err != nil {
			fmt.Println("Error when reading string data:", err)
			return
		}
		fmt.Println(string(strData))
	}
}

func readUintToUint64(r io.Reader, fi *gore.FileInfo) (uint64, error) {
	if fi.WordSize == 4 {
		var a uint32
		err := binary.Read(r, fi.ByteOrder, &a)
		return uint64(a), err
	}
	var a uint64
	err := binary.Read(r, fi.ByteOrder, &a)
	return a, err
}

func srcLineInfo(r2 *r2g2.Client, file *gore.GoFile) {
	fn, err := r2.GetCurrentFunction()
	if err != nil {
		fmt.Printf("Failed to get current function: %s.\n", err)
		return
	}

	tbl, err := file.PCLNTab()
	if err != nil {
		fmt.Printf("Failed to get lookup table: %s.\n", err)
		return
	}

	var curFile string
	var curLine int
	for _, pc := range fn.Ops {
		fileStr, line, _ := tbl.PCToLine(pc.Offset)

		// Check if on the same source line.
		if line == curLine && fileStr == curFile {
			continue
		}
		curLine = line
		curFile = fileStr

		// Add line as multiline comment.
		comment := fmt.Sprintf("%s:%d", fileStr, line)
		encodedComment := base64.StdEncoding.EncodeToString([]byte(comment))

		// Execute command.
		cmd := fmt.Sprintf("CCu base64:%s @ 0x%x", encodedComment, pc.Offset)
		_, err := r2.Run(cmd)
		if err != nil {
			fmt.Println("Error when adding comment:", err)
			return
		}
	}
}
