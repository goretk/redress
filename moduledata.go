// Copyright 2021 The GoRE Authors. All rights reserved.
// Use of this source code is governed by the license that
// can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/cheynewallace/tabby"
	"github.com/goretk/gore"
	"github.com/spf13/cobra"
)

func init() {
	modCMD := &cobra.Command{
		Use:     "moduledata path/to/file",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"md"},
		Short:   "Display sections extracted from the moduledata structure.",
		Run: func(cmd *cobra.Command, args []string) {
			moduledata(args[0])
		},
	}

	dumpCMD := &cobra.Command{
		Use:   "dump section path/to/file",
		Args:  cobra.ExactValidArgs(2),
		Short: "Dump the contents of a section to standard out.",
		Run: func(cmd *cobra.Command, args []string) {
			dumpModuleSection(args[1], args[0])
		},
	}
	modCMD.AddCommand(dumpCMD)

	rootCmd.AddCommand(modCMD)
}

func dumpModuleSection(fp, section string) {
	f, err := gore.Open(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open the file: %s.\n", err)
		return
	}
	defer f.Close()

	md, err := f.Moduledata()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not retrieve the file's moduledata: %s.\n", err)
		return
	}

	var sec gore.ModuleDataSection

	switch section {
	case "text":
		sec = md.Text()
	case "types":
		sec = md.Types()
	case "itablinks":
		sec = md.ITabLinks()
	case "pclntab":
		sec = md.PCLNTab()
	case "functab":
		sec = md.FuncTab()
	case "noptrdata":
		sec = md.NoPtrData()
	case "data":
		sec = md.Data()
	case "bss":
		sec = md.Bss()
	case "noptrbss":
		sec = md.NoPtrBss()
	default:
		fmt.Fprintf(os.Stderr, "No known section with the name %s.\n", section)
		return
	}

	data, err := sec.Data()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when getting the secton data: %s.\n", err)
		return
	}
	if len(data) == 0 {
		fmt.Fprintf(os.Stderr, "Section %s is empty.\n", section)
		return
	}

	r := bytes.NewReader(data)
	io.Copy(os.Stdout, r)
}

func moduledata(fp string) {
	f, err := gore.Open(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open the file: %s.\n", err)
		return
	}
	defer f.Close()

	md, err := f.Moduledata()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not retrieve the file's moduledata: %s.\n", err)
		return
	}

	tab := tabby.New()

	tab.AddHeader("Section", "Address", "Size")

	sections := []struct {
		name string
		info gore.ModuleDataSection
	}{
		{"text", md.Text()},
		{"types", md.Types()},
		{"itablinks", md.ITabLinks()},
		{"pclntab", md.PCLNTab()},
		{"functab", md.FuncTab()},
		{"noptrdata", md.NoPtrData()},
		{"data", md.Data()},
		{"bss", md.Bss()},
		{"noptrbss", md.NoPtrBss()},
	}

	for _, v := range sections {
		tab.AddLine(v.name, fmt.Sprintf("0x%x", v.info.Address), fmt.Sprintf("0x%x", v.info.Length))
	}

	tab.Print()
}
