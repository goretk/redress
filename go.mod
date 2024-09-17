module github.com/goretk/redress

go 1.21

toolchain go1.22.5

require (
	github.com/TcM1911/r2g2 v0.3.2
	github.com/cheynewallace/tabby v1.1.1
	github.com/goretk/gore v0.10.0
	github.com/spf13/cobra v1.2.1
)

require (
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/arch v0.7.0 // indirect
)

// This is used during development and disabled for release builds.
replace github.com/goretk/gore => ./gore
