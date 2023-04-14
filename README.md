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
(`f94ca9b1b01a7b06f19afaac3fbe0a43075c775a`) will be used.
The malware was first reported by
[Malwarebytes](https://blog.malwarebytes.com/threat-analysis/mac-threat-analysis/2018/05/new-mac-cryptominer-uses-xmrig/).

## Running it standalone

To run *redress*, just execute it on the command line. Below are some of the
possible flags that can be given. It is possible to use multiple flags to
extract different data. If no flags are given, no data is extracted. The idea
is to print more information than what is asked by the user.

```
% redress -h    
______         _                  
| ___ \       | |                 
| |_/ /___  __| |_ __ ___ ___ ___ 
|    // _ \/ _  | '__/ _ / __/ __|
| |\ |  __| (_| | | |  __\__ \__ \
\_| \_\___|\__,_|_|  \___|___|___/

Usage:
  redress [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  gomod       Display go mod information.
  help        Help about any command
  info        Print summary information.
  moduledata  Display sections extracted from the moduledata structure.
  packages    List packages.
  r2          Use redress with in r2.
  source      Source Code Projection.
  types       List types.
  version     Display redress version information.

Flags:
  -h, --help   help for redress

Use "redress [command] --help" for more information about a command.
```

### Information

Information about Go binary is shown by using the `info` command. Information
displayed include the compiler version and the date it was released, the build
id, GoRoot, and the folder path for the main package. 
```
% redress info pplauncher      
OS         macOS
Arch       amd64
Compiler   1.10 (2018-02-16)
Build ID   3xkvDz8awVOl0TA6jJs9/tcgWkdP85A-MJ6hpnmKm/_xkzRloiGp8H5mHsQgkh/7Kj4aF9BTT0MY_gaQnQI
GoRoot     /usr/local/Cellar/go/1.10/libexec
Main root  /Users/ronald/git/go-workspace/src/keybase.io/safetycrew/pplauncher
# main     1
# std      35
# vendor   2
```

### Packages

The different Go packages used in the binary can be extracted with the `packages`
command. *Redress* tries to only print the packages that are part of the project
and skips standard library and 3rd party library packages.

```
% redress packages pplauncher 
Packages:
Name  Version
----  -------
main  
```

Sometimes though, *redress* fails to classify a package. In this case, the
unclassified packages can be printed by also provide the `--unknown` flag:

```
% redress packages pplauncher --unknown
Packages:
Name  Version
----  -------
main  

Unknown Packages:
Name  Version
----  -------
```
To also include the standard library, use the `--std` flag. For 3rd party
packages, use the flag `--vendor`.

```
% redress packages pplauncher --std --vendor
Packages:
Name  Version
----  -------
main  

Vendors:
Name                                                                        Version
----                                                                        -------
vendor/golang_org/x/net/route                                               
vendor/golang_org/x/net/route.(*wireFormat).(vendor/golang_org/x/net/route  

Standard Library Packages:
Name                   Version
----                   -------
bufio                  
bytes                  
compress/flate         
compress/gzip          
context                
encoding/binary        
errors                 
fmt                    
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

The folder locations can be included in the output by using the `--filepath` flag.

```
% redress packages pplauncher --std --filepath 
Packages:
Name  Version  Path
----  -------  ----
main           /Users/ronald/git/go-workspace/src/keybase.io/safetycrew/pplauncher

Standard Library Packages:
Name                   Version  Path
----                   -------  ----
bufio                           /usr/local/Cellar/go/1.10/libexec/src/bufio
bytes                           /usr/local/Cellar/go/1.10/libexec/src/runtime
compress/flate                  /usr/local/Cellar/go/1.10/libexec/src/compress/flate
compress/gzip                   /usr/local/Cellar/go/1.10/libexec/src/compress/gzip
context                         /usr/local/Cellar/go/1.10/libexec/src/context
encoding/binary                 .
errors                          /usr/local/Cellar/go/1.10/libexec/src/errors
fmt                             /usr/local/Cellar/go/1.10/libexec/src/fmt
hash                            .
hash/crc32                      /usr/local/Cellar/go/1.10/libexec/src/hash/crc32
internal/cpu                    /usr/local/Cellar/go/1.10/libexec/src/internal/cpu
internal/poll                   /usr/local/Cellar/go/1.10/libexec/src/runtime
internal/singleflight           /usr/local/Cellar/go/1.10/libexec/src/internal/singleflight
internal/testlog                /usr/local/Cellar/go/1.10/libexec/src/internal/testlog
io                              /usr/local/Cellar/go/1.10/libexec/src/io
io/ioutil                       .
math                            .
math/rand                       /usr/local/Cellar/go/1.10/libexec/src/math/rand
net                             /usr/local/Cellar/go/1.10/libexec/src/net
os                              /usr/local/Cellar/go/1.10/libexec/src/runtime
os/exec                         /usr/local/Cellar/go/1.10/libexec/src/os/exec
os/signal                       /usr/local/Cellar/go/1.10/libexec/src/runtime
path/filepath                   /usr/local/Cellar/go/1.10/libexec/src/path/filepath
reflect                         /usr/local/Cellar/go/1.10/libexec/src/runtime
runtime                         /usr/local/Cellar/go/1.10/libexec/src/runtime
runtime/debug                   /usr/local/Cellar/go/1.10/libexec/src/runtime
sort                            /usr/local/Cellar/go/1.10/libexec/src/sort
strconv                         /usr/local/Cellar/go/1.10/libexec/src/strconv
strings                         /usr/local/Cellar/go/1.10/libexec/src/runtime
sync                            /usr/local/Cellar/go/1.10/libexec/src/runtime
sync/atomic                     /usr/local/Cellar/go/1.10/libexec/src/sync/atomic
syscall                         /usr/local/Cellar/go/1.10/libexec/src/runtime
time                            /usr/local/Cellar/go/1.10/libexec/src/runtime
unicode                         /usr/local/Cellar/go/1.10/libexec/src/unicode
unicode/utf8                    /usr/local/Cellar/go/1.10/libexec/src/unicode/utf8
```

### Extracting types

*Redress* has multiple commands that can be used to extract different type data.
Interfaces can be extracted with the `interface` command.  By default, standard
library and vendor interfaces are filtered out. These can be included by also
providing the `--std` and `--vendor` flag.

```
% redress types interface pplauncher  
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

Structures can be extracted with the `struct` command. Same as with interfaces,
standard library and vendor structures are filtered out but can be included by
providing the `--std` or `--vendor` flag.

```
% redress types struct pplauncher
...
type main.bintree struct {
	Func     func() (*main.asset, error)
	Children map[string]*main.bintree
}
...

type main.asset struct {
	bytes []uint8
	info  os.FileInfo
}

type main.bindataFileInfo struct {
	name    string
	size    int64
	mode    uint32
	modTime time.Time
}
...
```

Methods definitions for the structure is shown by including the `--methods` flag.

```
% redress types struct pplauncher --methods
...
type main.asset struct {
	bytes []uint8
	info  os.FileInfo
}

type main.bindataFileInfo struct {
	name    string
	size    int64
	mode    uint32
	modTime time.Time
}
func (main.bindataFileInfo) IsDir() bool
func (main.bindataFileInfo) ModTime() time.Time
func (main.bindataFileInfo) Mode() uint32
func (main.bindataFileInfo) Name() string
func (main.bindataFileInfo) Size() int64
func (main.bindataFileInfo) Sys() interface {}
...
```

It is also possible to print all types in the binary by using the `all` command.

### Estimating source code layout

One feature of _redress_ is to reconstruct the source code tree layout. This
can be done by using the `source` command. By default, standard library and 3rd
party packages are excluded but can be included by providing the flags `--std`,
`--vendor`, and/or `--unknown` flags.

The output includes the package name and its folder location at compile time.
For each file, the functions defined within are printed. The output also
includes auto generated functions produced by the compiler. For each function,
*redress* tries to guess the starting and ending line number.

```
% redress source pplauncher 
Package main: /Users/ronald/git/go-workspace/src/keybase.io/safetycrew/pplauncher
File: <autogenerated>	
	init Lines: 1 to 1 (0)	
	(*bindataFileInfo)Mode Lines: 1 to 1 (0)	
	(*bindataFileInfo)Name Lines: 1 to 1 (0)	
	(*bindataFileInfo)ModTime Lines: 1 to 1 (0)	
	(*bindataFileInfo)Size Lines: 1 to 1 (0)	
	(*bindataFileInfo)Sys Lines: 1 to 1 (0)	
	(*bindataFileInfo)IsDir Lines: 1 to 1 (0)	
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
	main Lines: 169 to 178 (9)	
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
% r2 pplauncher 
 -- radare2 contributes to the One Byte Per Child foundation.
[0x01055c90]> #!pipe redress r2 init
Compiler version: go1.10 (2018-02-16T16:05:53Z)
39 packages found.
2943 function symbols found
Analyzing all init functions.
Analyzing all main.main.
1717 type symbols found
```

_Redress_ also support some flags when executed from within _radare2_. These
flags can be used to print the Go definition for a specific type.

```
[0x01055c90]> #!pipe redress r2 help
Use redress with in r2.

Usage:
  redress r2 [command]

Aliases:
  r2, radare, radare2, r

Available Commands:
  init        Perform the initial analysis
  line        Annotate function with source lines.
  strarr      Print string array.
  type        Print type definition.

Flags:
  -h, --help   help for r2

Use "redress r2 [command] --help" for more information about a command.
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

By executing _redress_ with the `type` command and the address, the type
definition is returned:

```
> #!pipe redress r2 type 0x01116c20
type main.bintree struct{
	Func func() (*main.asset, error)
	Children map[string]*main.bintree
}
```

It is possible to chain the two commands:

```
:> #!pipe redress r2 type `f~sym.type.main.bintree~[0]`
type main.bintree struct{
	Func func() (*main.asset, error)
	Children map[string]*main.bintree
}
```

A better way is to define a _radare2_ macro:

```
:> (type flag,#!pipe redress r2 type `f~$0~[0]`)
:> .(type sym.type.main.bintree)
type main.bintree struct{
	Func func() (*main.asset, error)
	Children map[string]*main.bintree
}
```
