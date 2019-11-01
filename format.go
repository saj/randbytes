package main

import (
	"encoding/hex"
	"io"
	"sort"
)

type formatter func(io.Writer, io.Reader) (n int64, err error)

type formatDescriptor struct {
	F    formatter
	Help string
}

const defaultFormat = "hex"

var formats = map[string]formatDescriptor{
	"eschex": formatDescriptor{
		F:    formatEscHex,
		Help: "sequence of \\x__ for each input byte where _ is in range [0-f]",
	},
	"hex": formatDescriptor{
		F:    formatHex,
		Help: "sequence of __ for each input byte where _ is in range [0-f]",
	},
	"hexdump": formatDescriptor{
		F:    formatHexdump,
		Help: "canonical hex+ASCII",
	},
	"null": formatDescriptor{
		F:    formatNull,
		Help: "disable formatting; random data is copied verbatim to output",
	},
}

var formatNames []string

func init() {
	formatNames = make([]string, len(formats))
	{
		i := 0
		for k := range formats {
			formatNames[i] = k
			i++
		}
	}
	sort.Strings(formatNames)
}

type byteFormatter func(io.Writer, byte) (n int, err error)

func copyBytes(w io.Writer, r io.Reader, fmt byteFormatter) (int64, error) {
	n := int64(0)
	b := make([]byte, 512)
	for {
		rn, rerr := r.Read(b)
		if rn > 0 {
			for i := 0; i < rn; i++ {
				wn, werr := fmt(w, b[i])
				n += int64(wn)
				if werr != nil {
					return n, werr
				}
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return n, rerr
		}
	}
	return n, nil
}

const hextable = "0123456789abcdef"

func formatEscHex(w io.Writer, r io.Reader) (int64, error) {
	return copyBytes(w, r, func(w io.Writer, b byte) (int, error) {
		return w.Write([]byte{'\\', 'x', hextable[b>>4], hextable[b&0x0f]})
	})
}

func formatHex(w io.Writer, r io.Reader) (int64, error) {
	return copyBytes(w, r, func(w io.Writer, b byte) (int, error) {
		return w.Write([]byte{hextable[b>>4], hextable[b&0x0f]})
	})
}

func formatHexdump(w io.Writer, r io.Reader) (int64, error) {
	dw := hex.Dumper(w)
	defer dw.Close()
	return io.Copy(dw, r)
}

func formatNull(w io.Writer, r io.Reader) (int64, error) {
	return io.Copy(w, r)
}
