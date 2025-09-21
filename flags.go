package main

import (
	"flag"
)

type FlagOptions struct {
	address string
	root    string
}

var Flags FlagOptions

func ParseFlags() {
	address := flag.String("address", "0.0.0.0:80", "address:port")
	root := flag.String("root", ".", "path to dir location")

	flag.Parse()

	Flags = FlagOptions{
		address: *address,
		root:    *root,
	}
}
