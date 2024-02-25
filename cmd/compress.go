package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"strings"

	"github.com/Avik32223/compress/internal/compression"
	"github.com/Avik32223/compress/internal/encoding"
)

func compress(reader *bufio.Reader, writer *bufio.Writer) {
	var s strings.Builder
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		s.WriteRune(r)
	}
	source := s.String()
	compressor := compression.Compressor[string]{
		Source: source,
		CompressorOpts: compression.CompressorOpts[string]{
			Engine: &encoding.HuffmanCode{},
		},
	}

	cdata, err := compressor.Compress()
	if err != nil {
		panic(err)
	}

	if _, err := writer.Write(cdata); err != nil {
		panic(err)
	}
	writer.Flush()
}

func decompress(reader *bufio.Reader, writer *bufio.Writer) {
	var s strings.Builder
	for {
		r, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		s.WriteByte(r)
	}
	source := s.String()
	compressor := compression.Compressor[string]{
		Source: source,
		CompressorOpts: compression.CompressorOpts[string]{
			Engine: &encoding.HuffmanCode{},
		},
	}

	cdata, err := compressor.Decompress()
	if err != nil {
		panic(err)
	}

	writer.Write(cdata)
	writer.Flush()
}

func main() {
	var compressFlag bool
	var decompressFlag bool
	flag.BoolVar(&compressFlag, "c", false, "compress")
	flag.BoolVar(&decompressFlag, "d", false, "decompress")
	flag.Parse()
	if compressFlag && decompressFlag {
		flag.PrintDefaults()
		os.Exit(1)
	}
	args := flag.Args()
	if !compressFlag && !decompressFlag {
		compressFlag = true
	}

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	if len(args) >= 1 {
		f, err := os.OpenFile(args[0], os.O_RDONLY, 0)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		reader = bufio.NewReader(f)
	}

	if len(args) >= 2 {
		f, err := os.OpenFile(args[1], os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		writer = bufio.NewWriter(f)
	}

	if compressFlag {
		compress(reader, writer)
	} else if decompressFlag {
		decompress(reader, writer)
	}
	os.Exit(0)
}
