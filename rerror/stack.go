package rerror

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"runtime/debug"

	"github.com/maruel/panicparse/v2/stack"
)

func Stack() string {
	b := new(bytes.Buffer)
	PretifyStack(bytes.NewReader(debug.Stack()), b)
	return b.String()
}

func PretifyStack(r io.Reader, w io.Writer) {
	s, suffix, err := stack.ScanSnapshot(r, w, stack.DefaultOpts())
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	// Find out similar goroutine traces and group them into buckets.
	buckets := s.Aggregate(stack.AnyValue).Buckets

	// Calculate alignment.
	srcLen := 0
	pkgLen := 0
	for _, bucket := range buckets {
		for _, line := range bucket.Signature.Stack.Calls {
			if l := len(fmt.Sprintf("%s:%d", line.SrcName, line.Line)); l > srcLen {
				srcLen = l
			}
			if l := len(filepath.Base(line.Func.ImportPath)); l > pkgLen {
				pkgLen = l
			}
		}
	}

	for _, bucket := range buckets {
		// Print the goroutine header.
		extra := ""
		if s := bucket.SleepString(); s != "" {
			extra += " [" + s + "]"
		}
		if bucket.Locked {
			extra += " [locked]"
		}

		if len(bucket.CreatedBy.Calls) != 0 {
			extra += fmt.Sprintf(" [Created by %s.%s @ %s:%d]", bucket.CreatedBy.Calls[0].Func.DirName, bucket.CreatedBy.Calls[0].Func.Name, bucket.CreatedBy.Calls[0].SrcName, bucket.CreatedBy.Calls[0].Line)
		}
		fmt.Fprintf(w, "%d: %s%s\n", len(bucket.IDs), bucket.State, extra)

		// Print the stack lines.
		for _, line := range bucket.Stack.Calls {
			fmt.Fprintf(
				w,
				"    %-*s %-*s %s(%s)\n",
				pkgLen, line.Func.DirName, srcLen,
				fmt.Sprintf("%s:%d", line.SrcName, line.Line),
				line.Func.Name, &line.Args)
		}
		if bucket.Stack.Elided {
			_, _ = io.WriteString(w, "    (...)\n")
		}
	}

	// If there was any remaining data in the pipe, dump it now.
	if len(suffix) != 0 {
		_, _ = w.Write(suffix)
	}
	if err == nil {
		_, _ = io.Copy(w, r)
	}
}
