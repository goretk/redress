# Redress - A tool for analyzing stripped binaries

The *redress* software is a tool for analyzing stripped Go binaries compiled
with the Go compiler. It extracts data from the binary and uses it to
reconstruct symbols and performs analysis. It essentially tries to "re-dress" a
"stripped" binary. It can be downloaded from its [GitHub
page](https://github.com/goretk/redress).

It has two operation modes. The first is a standalone mode where the binary is
executed on its own. The second mode is used when the binary is executed from
within *radare2* via *r2pipe*. The binary is aware of its environment and
behaves accordingly.

For the examples shown the malware __pplauncher__
(`f94ca9b1b01a7b06f19afaac3fbe0a43075c775a`) will be used. A sample of the
binary can be downloaded [here](https://keybase.pub/joakimkennedy/gomalware/).
The malware was first reported by
[Malwarebytes](https://blog.malwarebytes.com/threat-analysis/mac-threat-analysis/2018/05/new-mac-cryptominer-uses-xmrig/).

## Running it standalone

To run *redress*, just execute it on the command line. Below are some of the
possible flags that can be given. It is possible to use multiple flags to
extract different data. If no flags are given, no data is extracted. The idea
is to print more information than what is asked by the user.

```
$ redress -h
Usage of redress:
  -compiler
    	Print information
  -filepath
    	Include file path for packages
  -interface
    	Print interfaces
  -method
    	Print type's methods
  -pkg
    	List packages
  -src
    	Print source tree
  -std
    	Include standard library packages
  -struct
    	Print structs
  -type
    	Print all type information
  -unknown
    	Include unknown packages
  -vendor
    	Include vendor package
  -version
    	Print redress version
```

### Packages

The different Go packages used in the binary can be extracted with the `-pkg`
flag. *Redress* tries to only print the packages that are part of the project
and skips standard library and 3rd party library packages.

```
$ redress -pkg pplauncher
Packages:
main
```

Sometimes though, *redress* fails to classify a package. In this case, the
unclassified packages can be printed by also provide the `-unknown` flag:

```
$ redress -pkg -unknown pplauncher
Packages:
main

Unknown Libraries:

```
To also include the standard library, use the `-std` flag. For 3rd party
packages, use the flag `-vendor`.

```
$ redress -pkg -vendor -std pplauncher
Packages:
main

Vendors:
vendor/golang_org/x/net/route
vendor/golang_org/x/net/route.(*wireFormat).(vendor/golang_org/x/net/route

Standard Libraries:
bufio
bytes
compress/flate
compress/gzip
context
encoding/binary
errors
fmt
go
hash
hash/crc32
internal/cpu
internal/poll
internal/singleflight
internal/testlog
io
io/ioutil
math
math/rand
net
os
os/exec
os/signal
path/filepath
reflect
runtime
runtime/debug
sort
strconv
strings
sync
sync/atomic
syscall
time
unicode
unicode/utf8
```

The folder location for the package can also be included by using the
`-filepath` flag.

```
$ redress -pkg -std -filepath pplauncher
Packages:
main | /Users/ronald/git/go-workspace/src/keybase.io/safetycrew/pplauncher

Standard Libraries:
bufio | /usr/local/Cellar/go/1.10/libexec/src/bufio
bytes | /usr/local/Cellar/go/1.10/libexec/src/runtime
compress/flate | /usr/local/Cellar/go/1.10/libexec/src/compress/flate
compress/gzip | /usr/local/Cellar/go/1.10/libexec/src/compress/gzip
context | /usr/local/Cellar/go/1.10/libexec/src/context
encoding/binary | .
errors | /usr/local/Cellar/go/1.10/libexec/src/errors
fmt | /usr/local/Cellar/go/1.10/libexec/src/fmt
go | .
hash | .
hash/crc32 | /usr/local/Cellar/go/1.10/libexec/src/hash/crc32
internal/cpu | /usr/local/Cellar/go/1.10/libexec/src/internal/cpu
internal/poll | /usr/local/Cellar/go/1.10/libexec/src/runtime
internal/singleflight | /usr/local/Cellar/go/1.10/libexec/src/internal/singleflight
internal/testlog | /usr/local/Cellar/go/1.10/libexec/src/internal/testlog
io | /usr/local/Cellar/go/1.10/libexec/src/io
io/ioutil | /usr/local/Cellar/go/1.10/libexec/src/io/ioutil
math | .
math/rand | /usr/local/Cellar/go/1.10/libexec/src/math/rand
net | /usr/local/Cellar/go/1.10/libexec/src/net
os | /usr/local/Cellar/go/1.10/libexec/src/runtime
os/exec | /usr/local/Cellar/go/1.10/libexec/src/os/exec
os/signal | /usr/local/Cellar/go/1.10/libexec/src/runtime
path/filepath | /usr/local/Cellar/go/1.10/libexec/src/path/filepath
reflect | /usr/local/Cellar/go/1.10/libexec/src/runtime
runtime | /usr/local/Cellar/go/1.10/libexec/src/runtime
runtime/debug | /usr/local/Cellar/go/1.10/libexec/src/runtime
sort | /usr/local/Cellar/go/1.10/libexec/src/sort
strconv | /usr/local/Cellar/go/1.10/libexec/src/strconv
strings | /usr/local/Cellar/go/1.10/libexec/src/runtime
sync | /usr/local/Cellar/go/1.10/libexec/src/runtime
sync/atomic | /usr/local/Cellar/go/1.10/libexec/src/sync/atomic
syscall | /usr/local/Cellar/go/1.10/libexec/src/runtime
time | /usr/local/Cellar/go/1.10/libexec/src/runtime
unicode | /usr/local/Cellar/go/1.10/libexec/src/unicode
unicode/utf8 | /usr/local/Cellar/go/1.10/libexec/src/unicode/utf8
```

### Compiler information

Information about the Go compiler used to build the binary can be
shown by using the `-compiler` flag. It prints the release version
and the time stamp when the release tag was created in the git tree.

```
$ redress -compiler pplauncher
Compiler version: go1.10 (2018-02-16T16:05:53Z)
```

### Extracting types

*Redress* has multiple flags that can be used to extract different type data.
Interfaces can be extracted with the `-interface` flag.  By default, standard
library interfaces are filtered out. These can be included by also providing
the `-std` flag.

```
$ redress -interface pplauncher
type error interface {
	Error() string
}

type interface {} interface{}

type route.Addr interface {
	Family() int
}

type route.Message interface {
	Sys() []route.Sys
}

type route.Sys interface {
	SysType() int
}

type route.binaryByteOrder interface {
	PutUint16([]uint8, uint16)
	PutUint32([]uint8, uint32)
	Uint16([]uint8) uint16
	Uint32([]uint8) uint32
	Uint64([]uint8) uint64
}
```

Structures can be extracted with the `-struct` flag (`redress -struct
pplauncher`). Same as for interfaces, standard library structures are filtered
out but can be included by also providing the `-std` flag.

```
type main.asset struct{
	bytes []uint8
	info os.FileInfo
}

type main.bindataFileInfo struct{
	name string
	size int64
	mode uint32
	modTime time.Time
}

type main.bintree struct{
	Func func() (*main.asset, error)
	Children map[string]*main.bintree
}
```

Methods for the structure can be shown with the command `redress -struct
-method pplauncher`.

```
type main.asset struct{
	bytes []uint8
	info os.FileInfo
}

type main.bindataFileInfo struct{
	name string
	size int64
	mode uint32
	modTime time.Time
}
func (main.bindataFileInfo) IsDir() bool
func (main.bindataFileInfo) ModTime() time.Time
func (main.bindataFileInfo) Mode() uint32
func (main.bindataFileInfo) Name() string
func (main.bindataFileInfo) Size() int64
func (main.bindataFileInfo) Sys() interface {}

type main.bintree struct{
	Func func() (*main.asset, error)
	Children map[string]*main.bintree
}
```

It is also possible to print all types in the binary by using the `-type` flag.
The `-method` flag can also be used to include defined methods.

### Estimating source code layout

One feature of _redress_ is to reconstruct the source code tree layout.  This
can be done by using the `-src` flag. By default, standard library and 3rd
party packages are excluded but can be included by providing the flags `-std`,
`-vendor`, and/or `-unknown`.

The output includes the package name and its folder location at compile time.
For each file, the functions defined within are printed. The output also
includes auto generated functions produced by the compiler. For each function,
*redress* tries to guess the starting and ending line number.

```
$ redress -src pplauncher
Package main: /Users/ronald/git/go-workspace/src/keybase.io/safetycrew/pplauncher
File: <autogenerated>
	init Lines: 1 to 164 (163)
	(*bindataFileInfo)Name Lines: 1 to 54 (53)
	(*bindataFileInfo)Size Lines: 1 to 57 (56)
	(*bindataFileInfo)Mode Lines: 1 to 60 (59)
	(*bindataFileInfo)ModTime Lines: 1 to 63 (62)
	(*bindataFileInfo)IsDir Lines: 1 to 1 (0)
	(*bindataFileInfo)Sys Lines: 1 to 1 (0)
File: bindata.go
	bindataRead Lines: 21 to 54 (33)
	bindataFileInfoName Lines: 54 to 57 (3)
	bindataFileInfoSize Lines: 57 to 60 (3)
	bindataFileInfoMode Lines: 60 to 63 (3)
	bindataFileInfoModTime Lines: 63 to 66 (3)
	bindataFileInfoIsDir Lines: 66 to 69 (3)
	bindataFileInfoSys Lines: 69 to 74 (5)
	dataLibmicrohttpd12DylibBytes Lines: 74 to 81 (7)
	dataLibmicrohttpd12Dylib Lines: 81 to 94 (13)
	dataMshelperBytes Lines: 94 to 101 (7)
	dataMshelper Lines: 101 to 115 (14)
	Asset Lines: 115 to 124 (9)
File: pplauncher.go
	init0 Lines: 17 to 21 (4)
	check Lines: 21 to 29 (8)
	cleanupMinerDirectory Lines: 29 to 52 (23)
	extractPayload Lines: 52 to 74 (22)
	fetchConfig Lines: 74 to 120 (46)
	launchMiner Lines: 120 to 142 (22)
	exitGracefully Lines: 142 to 152 (10)
	autoKill Lines: 152 to 160 (8)
	autoKillfunc1 Lines: 153 to 163 (10)
	handleExit Lines: 160 to 169 (9)
	handleExitfunc1 Lines: 163 to 166 (3)
	main Lines: 169 to 175 (6)
```

## Using redress with radare2

_Redress_ can be executed from within _radare2_ via _r2pipe_. If _redress_ is
executed with no flags, it performs an automatic analysis of the binary.  It
will extract all the functions and methods and construct symbol flags for them.
It will also extract the types in the binary. It will mark the corresponding
`_type` with this flag to make analysis easier. By doing this, it highlights
which actual type is being allocated on the heap for easier analysis. The last
thing _redress_ does is to also tell _radare2_ to analyze the `main.init` and
`main.main` function recursively.

```
$ r2 pplauncher
 -- There's no way you could crash radare2. No. Way.
[0x01055c90]> #!pipe redress
Compiler version: go1.10 (2018-02-16T16:05:53Z)
40 packages found.
2659 function symbols found
1717 type symbols found
```

_Redress_ also support some flags when executed from within _radare2_. These
flags can be used to print the Go definition for a specific type.

```
[0x01055c90]> #!pipe redress -h
Usage of redress:
  -method
    	Print type's methods
  -type int
    	Lookup the Go definition for a type
  -version
    	Print redress version
```

### Working with types

The type identification by _redress_ makes the analysis easier. In the code
snippet below, it can be seen that memory for the type `main.bintree` is
to be allocated.

```
; CODE XREF from sym.main.init (0x10eb3de)
0x010eb14d      e80ed2f1ff     call sym.runtime.makemap_small
0x010eb152      488b0424       mov rax, qword [rsp]
0x010eb156      4889442440     mov qword [var_40h], rax
0x010eb15b      488d0dbeba02.  lea rcx, sym.type.main.bintree
0x010eb162      48890c24       mov qword [rsp], rcx
0x010eb166      e8155ef2ff     call sym.runtime.newobject
```

To get the type definition of this type, the address for the flag needs to be
known.

```
:> f~sym.type.main.bintree
0x01116c20 1 sym.type.main.bintree
```

By executing _redress_ with the `-type` flag and the address, the type definition
is returned:

```
:> #!pipe redress -type 0x01116c20
type main.bintree struct{
	Func func() (*main.asset, error)
	Children map[string]*main.bintree
}
```

It is possible to chain the two commands:

```
:> #!pipe redress -type `f~sym.type.main.bintree~[0]`
type main.bintree struct{
	Func func() (*main.asset, error)
	Children map[string]*main.bintree
}
```

A better way is to define a _radare2_ macro:

```
:> (type flag,#!pipe redress -type `f~$0~[0]`)
:> .(type sym.type.main.bintree)
type main.bintree struct{
	Func func() (*main.asset, error)
	Children map[string]*main.bintree
}
```

The methods for the type can also be included in the output by including the
`-method` flag. Below a new macro is defined.

```
(type+ flag,#!pipe redress -method -type `f~$0~[0]`)
```

In the code snippet below, it can be seen that memory for the `exec.Cmd`
structure is being allocated.

```
0x010e5cc5      488d0514b804.  lea rax, sym.type.exec.Cmd
0x010e5ccc      48890424       mov qword [rsp], rax
0x010e5cd0      e8abb2f2ff     call sym.runtime.newobject
```

The `exec.Cmd` does not have any methods because they are associated with the
pointer type (`*exec.Cmd`) for the structure. To get the methods, this type has
to be used. It can be found by grepping the flags for `exec.Cmd`.  _Radare2_
will replace the "__*__" with a "**_**" in the flag name. Using `_exec.Cmd`
instead does return the methods.

```
:> .(type+ sym.type.exec.Cmd)
type exec.Cmd struct{
	Path string
	Args []string
	Env []string
	Dir string
	Stdin io.Reader
	Stdout io.Writer
	Stderr io.Writer
	ExtraFiles []*os.File
	SysProcAttr *syscall.SysProcAttr
	Process *os.Process
	ProcessState *os.ProcessState
	ctx context.Context
	lookPathErr error
	finished bool
	childFiles []*os.File
	closeAfterStart []io.Closer
	closeAfterWait []io.Closer
	goroutine []func() error
	errch chan error
	waitDone chan struct {}
}
:> f~exec.Cmd
0x010ff100 1 sym.type._struct___F_uintptr__pw__os.File__c__exec.Cmd
0x0111ada0 1 sym.type.struct___F_uintptr__pw__os.File__c__exec.Cmd
0x0112c280 1 sym.type._exec.Cmd
0x011314e0 1 sym.type.exec.Cmd
:> .(type+ sym.type._exec.Cmd)
*exec.Cmd
func (*exec.Cmd) CombinedOutput()
func (*exec.Cmd) Output()
func (*exec.Cmd) Run() error
func (*exec.Cmd) Start() error
func (*exec.Cmd) StderrPipe()
func (*exec.Cmd) StdinPipe()
func (*exec.Cmd) StdoutPipe()
func (*exec.Cmd) Wait() error
func (*exec.Cmd) argv()
func (*exec.Cmd) closeDescriptors()
func (*exec.Cmd) envv()
func (*exec.Cmd) stderr()
func (*exec.Cmd) stdin()
func (*exec.Cmd) stdout()
func (*exec.Cmd) writerDescriptor()
```
