package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/alecthomas/units"
)

func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(w, "stream-split [options] [files...] -- <external command>")
		flag.PrintDefaults()
	}

	// Determine start of the external command to split up the args.
	cmdIdx := -1
	for i, arg := range os.Args {
		if arg == "--" {
			cmdIdx = i
			break
		}
	}

	if cmdIdx == -1 {
		flag.Usage()
		return
	}

	cmdArgs := os.Args[cmdIdx+1:]
	os.Args = os.Args[:cmdIdx]

	var (
		maxLines    int
		maxBytesStr string
		maxBytes    int64
		debugLog    bool
	)

	flag.IntVar(&maxLines, "lines", 0, "Number of lines in a split.")
	flag.StringVar(&maxBytesStr, "bytes", "", "Maximum number of bytes to split on.")
	flag.BoolVar(&debugLog, "debug", false, "Turn on debug/verbose logging.")

	flag.Parse()

	var rdr io.Reader

	// Input files provided.
	if len(flag.Args()) > 0 {
		var rdrs []io.Reader

		for _, arg := range flag.Args() {
			f, err := os.Open(arg)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			rdrs = append(rdrs, f)
		}

		rdr = io.MultiReader(rdrs...)
	} else {
		rdr = os.Stdin
	}

	if maxBytesStr != "" {
		var err error
		b, err := units.ParseBase2Bytes(maxBytesStr)
		if err != nil {
			log.Fatal(err)
		}
		maxBytes = int64(b)
	}

	if maxLines == 0 && maxBytes == 0 {
		log.Fatal("The -lines or -bytes option must be provided.")
	}

	buf := bytes.NewBuffer(nil)

	// 5 MB
	scBufSize := 5 * 1024 * 1024
	sc := bufio.NewScanner(rdr)
	sc.Buffer(make([]byte, scBufSize), scBufSize)
	sc.Split(bufio.ScanLines)

	var count int
	splitNum := 1

	for sc.Scan() {
		data := sc.Bytes()
		currentBytes := buf.Len()

		// Flush before writing next.
		if maxBytes > 0 {
			totalBytes := currentBytes + len(data) + 1

			// Add one for re-adding the newline
			if int64(totalBytes) > maxBytes {
				if debugLog {
					log.Printf("--- split %d (%s; %d lines) ---", splitNum, units.Base2Bytes(currentBytes), count)
				}
				if err := runCmd(buf, cmdArgs); err != nil {
					log.Fatal(err)
				}

				count = 0
				splitNum++
			}
		}

		buf.Write(data)
		buf.WriteByte('\n')

		currentBytes = buf.Len()

		if maxLines > 0 {
			count++
			if count == maxLines {
				if debugLog {
					log.Printf("--- split %d (%s; %d lines) ---", splitNum, units.Base2Bytes(currentBytes), count)
				}
				if err := runCmd(buf, cmdArgs); err != nil {
					log.Fatal(err)
				}

				count = 0
				splitNum++
			}
		}
	}

	if buf.Len() > 0 {
		if debugLog {
			log.Printf("--- split %d (%s; %d lines) ---", splitNum, units.Base2Bytes(buf.Len()), count)
		}
		if err := runCmd(buf, cmdArgs); err != nil {
			log.Fatal(err)
		}
	}

	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
}

func runCmd(buf *bytes.Buffer, args []string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = buf
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	buf.Reset()
	return nil
}
