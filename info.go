// Copyright 2021 The GoRE Authors. All rights reserved.
// Use of this source code is governed by the license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cheynewallace/tabby"
	gore "github.com/goretk/gore"
	"github.com/spf13/cobra"
)

func init() {
	infoCMD := &cobra.Command{
		Use:     "info path/to/go/file",
		Aliases: []string{"metadata", "i"},
		Short:   "Print summary information.",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			listInfo(args[0])
		},
	}

	gomodCMD := &cobra.Command{
		Use:     "gomod /path/to/go/file",
		Aliases: []string{"gosum", "gm"},
		Short:   "Display go mod information.",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			listModInfo(args[0])
		},
	}

	rootCmd.AddCommand(gomodCMD)
	rootCmd.AddCommand(infoCMD)
}

func listInfo(fileStr string) {
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

	t := tabby.New()

	t.AddLine("OS", f.FileInfo.OS)
	t.AddLine("Arch", f.FileInfo.Arch)

	if comp, err := f.GetCompilerVersion(); err == nil {
		t.AddLine("Compiler", fmt.Sprintf("%s (%s)", strings.TrimPrefix(comp.Name, "go"), strings.Split(comp.Timestamp, "T")[0]))
	}

	if f.BuildID != "" {
		t.AddLine("Build ID", f.BuildID)
	}

	if root, err := f.GetGoRoot(); err == nil {
		t.AddLine("GoRoot", root)
	}

	if pkg, err := f.GetPackages(); err == nil {
		for _, p := range pkg {
			if p.Name == "main" {
				t.AddLine("Main root", p.Filepath)
				break
			}
		}
		t.AddLine("# main", len(pkg))

		std, _ := f.GetSTDLib()
		t.AddLine("# std", len(std))

		ven, _ := f.GetVendors()
		t.AddLine("# vendor", len(ven))

		unk, _ := f.GetUnknown()
		if len(unk) != 0 {
			t.AddLine("# unknown", len(unk))
		}
	}

	t.Print()
}

func listModInfo(fileStr string) {
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

	if f.BuildInfo == nil {
		fmt.Fprintf(os.Stderr, "No build info found in the file.\n")
		return
	}

	if f.BuildInfo.ModInfo == nil {
		fmt.Fprintf(os.Stderr, "No mod info found in the file.\n")
		return
	}

	mod := f.BuildInfo.ModInfo

	t := tabby.New()

	t.AddHeader("Type", "Name", "Version", "Replaced by", "Hash")

	// If the main path is empty, we fallback to the path in the
	// build info.
	mainPath := mod.Main.Path
	if mainPath == "" {
		mainPath = mod.Path
	}
	t.AddLine("main", mainPath, mod.Main.Version, "", mod.Main.Sum)

	for _, m := range mod.Deps {
		if m.Replace != nil {
			t.AddLine("dep", m.Path, m.Version, m.Replace.Path, m.Sum)
			t.AddLine("replacement", m.Replace.Path, m.Replace.Version, "", m.Sum)
		} else {
			t.AddLine("dep", m.Path, m.Version, "", m.Sum)
		}
	}
	t.Print()
	fmt.Printf("\n")
}
