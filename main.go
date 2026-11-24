package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

func main() {
	file := flag.String("file", "gnolive.gno", "path to the Gno source file")
	once := flag.Bool("once", false, "print the file once and exit")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, *file, *once, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "gnolive:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, path string, once bool, out, diagnostics io.Writer) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if once {
		source, err := prepare(path)
		if err != nil {
			return err
		}
		_, err = out.Write(source)
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Follow the directory because editors can replace the original inode.
	if err := watcher.Add(filepath.Dir(path)); err != nil {
		return err
	}

	var last []byte
	printChanged := func() error {
		source, err := prepare(path)
		if err != nil {
			fmt.Fprintln(diagnostics, "gnolive:", err)
			return nil
		}
		if bytes.Equal(source, last) {
			return nil
		}
		if _, err := out.Write(source); err != nil {
			return err
		}
		last = source
		return nil
	}

	if err := printChanged(); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Name == path && event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
				if err := printChanged(); err != nil {
					return err
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return err
		}
	}
}

func prepare(path string) ([]byte, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}
	var result bytes.Buffer
	if err := printer.Fprint(&result, fset, node); err != nil {
		return nil, err
	}
	result.WriteByte('\n')
	return result.Bytes(), nil
}
