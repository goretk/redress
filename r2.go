// Copyright 2019-2021 The GoRE Authors. All rights reserved.
// Use of this source code is governed by the license that
// can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/TcM1911/r2g2"
	gore "github.com/goretk/gore"
	"github.com/spf13/cobra"
)

func cleanupName(old string) string {
	replacer := strings.NewReplacer(
		" ", "_",
		"-", "_",
		";", "_",
		"/", "_",
		"@", "_",
	)
	return replacer.Replace(old)
}

func init() {
	r2Cmd := &cobra.Command{
		Use:     "r2",
		Aliases: []string{"radare", "radare2", "r"},
		Short:   "Use redress with in r2.",
		// Long:    longR2Help,
	}

	var useComment bool
	srcCMD := &cobra.Command{
		Use:   "line",
		Short: "Annotate function with source lines.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if !r2g2.CheckForR2Pipe() {
				fmt.Println("This command can only be executed from within radare2.")
				os.Exit(1)
			}
			annotateWithSourceLine(useComment)
		},
	}
	srcCMD.Flags().BoolVarP(&useComment, "comment", "c", false, "Use comments instead of flags.")

	r2Cmd.AddCommand(srcCMD)

	typCMD := &cobra.Command{
		Use:   "type",
		Short: "Print type definition.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !r2g2.CheckForR2Pipe() {
				fmt.Println("This command can only be executed from within radare2.")
				return
			}
			addr, err := strconv.ParseUint(args[0], 0, strconv.IntSize)
			if err != nil {
				fmt.Printf("Bad address format: %s.\n", err)
				return
			}
			resolveTypeAt(addr)
		},
	}
	r2Cmd.AddCommand(typCMD)

	strArrCMD := &cobra.Command{
		Use:   "strarr offset length",
		Short: "Print string array.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if !r2g2.CheckForR2Pipe() {
				fmt.Println("This command can only be executed from within radare2.")
				return
			}

			addr, err := strconv.ParseUint(args[0], 0, strconv.IntSize)
			if err != nil {
				fmt.Printf("Bad address format: %s.\n", err)
				return
			}

			length, err := strconv.ParseUint(args[1], 0, strconv.IntSize)
			if err != nil {
				fmt.Printf("Bad length format: %s.\n", err)
				return
			}

			extractStringSlice(addr, length)
		},
	}
	r2Cmd.AddCommand(strArrCMD)

	intCMD := &cobra.Command{
		Use:     "init",
		Short:   "Perform the initial analysis",
		Aliases: []string{"analyze", "aaa"},
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if !r2g2.CheckForR2Pipe() {
				fmt.Println("This command can only be executed from within radare2.")
				return
			}

			initAnal()
		},
	}
	r2Cmd.AddCommand(intCMD)

	rootCmd.AddCommand(r2Cmd)
}

func extractStringSlice(addr, length uint64) {
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

	printStringSlice(file, addr, length)
}

func resolveTypeAt(addr uint64) {
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

	lookupType(file, addr)
}

func annotateWithSourceLine(useComment bool) {
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
	srcLineInfo(r2, file, useComment)
}

func initAnal() {
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

	var correction uint64
	if file.FileInfo.OS == "windows" {
		textStart := getFileSectionAddress(r2, ".text")
		correction = findAddressCorrection(file, textStart)
		if correction != 0 {
			fmt.Printf("PE .text section and Go runtime mismatch. Using address correction 0x%x.\n", correction)
		}
	}

	// Vendors, stdlib and unknown have now been populated so we can ignore the err check.
	vendors, _ := file.GetVendors()
	std, _ := file.GetSTDLib()
	unknown, _ := file.GetUnknown()
	generated, _ := file.GetGeneratedPackages()

	pkgs = append(pkgs, vendors...)
	pkgs = append(pkgs, std...)
	pkgs = append(pkgs, unknown...)
	pkgs = append(pkgs, generated...)

	fmt.Printf("%d packages found.\n", len(pkgs))
	applyFuncSymbols(pkgs, r2, correction)

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

func findAddressCorrection(file *gore.GoFile, headerTextAddress uint64) uint64 {
	modData, err := file.Moduledata()
	if err != nil {
		return 0
	}
	mtxt := modData.Text().Address
	if headerTextAddress >= mtxt {
		return 0
	}
	return mtxt - headerTextAddress
}

func applyFuncSymbols(pkgs []*gore.Package, r2 *r2g2.Client, correction uint64) {
	count := 0
	for _, p := range pkgs {
		for _, f := range p.Functions {
			if f.Offset == uint64(0) {
				continue
			}
			r2.NewFlagWithLength(
				"fcn."+cleanupName(p.Name)+"."+cleanupName(f.Name),
				f.Offset+correction,
				f.End-f.Offset)
			count++
		}
		for _, m := range p.Methods {
			if m.Offset == uint64(0) {
				continue
			}
			r2.NewFlagWithLength(
				"fcn."+cleanupName(p.Name)+"#"+cleanupName(m.Receiver)+"."+cleanupName(m.Name),
				m.Offset+correction,
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

func srcLineInfo(r2 *r2g2.Client, file *gore.GoFile, useComment bool) {
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

		var cmd string
		if useComment {
			encodedComment := base64.StdEncoding.EncodeToString([]byte(comment))
			cmd = fmt.Sprintf("CCu base64:%s @ 0x%x", encodedComment, pc.Offset)
		} else {
			r2.Run("fs line")
			cmd = fmt.Sprintf("f %s 1 @ 0x%x", cleanupName(comment), pc.Offset)
		}

		// Execute command.
		_, err := r2.Run(cmd)
		if err != nil {
			fmt.Println("Error when adding comment:", err)
			return
		}
	}
	if !useComment {
		r2.Run("fs *")
	}
}

func getFileSectionAddress(r2 *r2g2.Client, name string) uint64 {
	data, err := r2.Run("iSj")
	if err != nil {
		return 0
	}

	var sections []struct {
		Name        string `json:"name"`
		Size        uint64 `json:"size"`
		VSize       uint64 `json:"vsize"`
		Permissions string `json:"perm"`
		PAddr       uint64 `json:"paddr"`
		VAddr       uint64 `json:"vaddr"`
	}

	err = json.Unmarshal(data, &sections)
	if err != nil {
		return 0
	}

	for _, s := range sections {
		if s.Name == name {
			return s.VAddr
		}
	}

	return 0
}
