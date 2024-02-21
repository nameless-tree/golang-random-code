package args

import (
	"flag"
	"os"
)

type Args struct {
	HttpAddr string
	Path     string
}

func ArgsParse() *Args {
	addr := flag.String("addr", "", "address of http server: `localhost:3333` for example")
	path := flag.String("path", "", "dir path to scan")

	flag.Parse()

	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if !flagset["addr"] || !flagset["path"] {
		flag.Usage()
		os.Exit(0)
	}

	return &Args{
		HttpAddr: *addr,
		Path:     *path,
	}
}
