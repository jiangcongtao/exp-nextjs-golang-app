package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// More verbose syntax to include all files, like . _ prefixed hidden files
// //go:embed out/*
// //go:embed out/_next/static/css/*
// //go:embed out/_next/static/chunks/*
// //go:embed out/_next/static/chunks/pages/*
// //go:embed out/_next/static/98eRoObwE0k4A3bguCFax/*
// var embeddedFS embed.FS

// More succinct synatx to include files, like . _ prefixed hidden files

//go:embed all:out
var embeddedFS embed.FS

type hybridFS struct {
	embedFS embed.FS
	root    string
}

const embedRootFolder = "out"

func (h *hybridFS) Open(name string) (fs.File, error) {
	path := filepath.Join(h.root, name)
	file, err := os.Open(path)

	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("Open embedded file: ", name)
		return h.embedFS.Open(filepath.Join(embedRootFolder, name))
	}
	fmt.Println("Open file from file system: ", path)
	return file, err
}

// cacheMiddleware sets the Cache-Control header for static files.
func cacheMiddleware(next http.Handler, maxAge time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age="+maxAge.String())
		next.ServeHTTP(w, r)
	})
}

func main() {
	exportFlag := flag.Bool("export", false, "Export embedded files to the specified folder and exit")
	physicalFSRootFlag := flag.String("physical-root", "external", "Path to the folder where files should be searched before embedded files")
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

	physicalFSRoot := *physicalFSRootFlag
	hfs := &hybridFS{
		embedFS: embeddedFS,
		root:    physicalFSRoot,
	}

	// distFS, err := fs.Sub(embeddedFS, "out")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Set the duration for which you want the cache to be valid
	// cacheDuration := 5 * time.Minute
	cacheDuration := time.Duration(0)

	http.HandleFunc("/process", processHandler)
	http.Handle("/", cacheMiddleware(http.FileServer(http.FS(hfs)), cacheDuration))
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

	return fs.WalkDir(embeddedFS, embedRootFolder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(embedRootFolder, path)
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
