package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	log.SetFlags(0)
}

func main() {
	formatHelp := func(_ *kingpin.ParseContext) error {
		for _, f := range formatNames {
			fmt.Printf("%10s\t%s\n", f, formats[f].Help)
		}
		os.Exit(0)
		return nil
	}

	var (
		app    = kingpin.New(filepath.Base(os.Args[0]), "Generate and format random data.")
		length = app.Arg("length", "Number of bytes to generate.").Required().Uint64()
		format = app.Flag("format", "Output format.  See --format-help.").Default(defaultFormat).Enum(formatNames...)
		_      = app.Flag("format-help", "List supported output formats and exit.").PreAction(formatHelp).Bool()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	pr, pw, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}
	eg := &errgroup.Group{}
	eg.Go(func() error {
		defer pw.Close()
		_, err := io.CopyN(pw, rand.Reader, int64(*length))
		return err
	})
	eg.Go(func() error {
		defer pr.Close()
		_, err := formats[*format].F(w, pr)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
