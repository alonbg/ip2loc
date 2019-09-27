package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tserkov/ip2loc"
)

func main() {
	ips := readArgs()

	if len(ips) == 0 {
		ips = readStdin()

		if len(ips) == 0 {
			usage()
		}
	}

	dbpath := os.Getenv("IP2LOC_DB")
	if dbpath == "" {
		fatal("Environment variable IP2LOC_DB is not set to the path of the IP2Location DB1 bin!")
	}

	db, err := ip2loc.New(dbpath)
	if err != nil {
		fatal(err)
	}

	for _, ip := range ips {
		r, err := db.Query(ip)
		if err != nil {
			fatal(err)
		}

		fmt.Fprintf(
			os.Stdout,
			"%-15s %s (%s)\n",
			ip,
			r.CountryName,
			r.CountryCode,
		)
	}
}

func readArgs() []string {
	if len(os.Args) == 1 {
		return []string{}
	}

	return os.Args[1:]
}

func readStdin() []string {
	info, err := os.Stdin.Stat()
	if err != nil {
		fatal(err)
	}

	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
		usage()
	}

	r := bufio.NewReader(os.Stdin)

	var bs []byte
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fatal(err)
			}
		}

		bs = append(bs, b)
	}

	return strings.Split(string(bs), " ")
}

func fatal(msg interface{}) {
	fmt.Fprintf(os.Stdout, "%s\n", msg)
	os.Exit(1)
}

func usage() {
	fatal("Usage: " + filepath.Base(os.Args[0]) + " ip_addr...")
}
