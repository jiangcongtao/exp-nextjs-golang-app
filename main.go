package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// More verbose syntax to include all files, like . _ prefixed hidden files
// //go:embed out/*
// //go:embed out/_next/static/css/*
// //go:embed out/_next/static/chunks/*
// //go:embed out/_next/static/chunks/pages/*
// //go:embed out/_next/static/98eRoObwE0k4A3bguCFax/*
// var embeddedFS embed.FS

// More succinct syntax to include files, like . _ prefixed hidden files

//go:embed all:out
var embeddedFS embed.FS

type hybridFS struct {
	embedFS embed.FS
	root    string
}

func (h *hybridFS) Open(name string) (fs.File, error) {
	path := filepath.Join(h.root, name)
	file, err := os.Open(path)

	if errors.Is(err, os.ErrNotExist) {
		return h.embedFS.Open(filepath.Join("static", name))
	}

	return file, err
}

func main() {
	exportFlag := flag.Bool("export", false, "Export embedded files to the specified folder and exit")
	flag.Parse()

	if *exportFlag {
		exportDir := flag.Arg(0)
		if exportDir == "" {
			fmt.Println("Please provide an export directory")
			os.Exit(1)
		}

		err := exportEmbeddedFiles(exportDir)
		if err != nil {
			fmt.Printf("Error exporting files: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Files exported to %s\n", exportDir)
		os.Exit(0)
	}

	distFS, err := fs.Sub(embeddedFS, "out")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/process", processHandler)
	http.Handle("/", http.FileServer(http.FS(distFS)))
	http.ListenAndServe(":8080", nil)
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/plain")
	w.Write(body)
}

func exportEmbeddedFiles(exportDir string) error {
	err := os.MkdirAll(exportDir, 0755)
	if err != nil {
		return err
	}

	return fs.WalkDir(embeddedFS, "out", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel("out", path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(exportDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		srcFile, err := embeddedFS.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
}
