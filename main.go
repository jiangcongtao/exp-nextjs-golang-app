package main

import (
	"embed"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
)

// More verbose syntax to include all files, like . _ prefixed hidden files
// //go:embed out/*
// //go:embed out/_next/static/css/*
// //go:embed out/_next/static/chunks/*
// //go:embed out/_next/static/chunks/pages/*
// //go:embed out/_next/static/98eRoObwE0k4A3bguCFax/*
// var content embed.FS

// More succinct synatx to include files, like . _ prefixed hidden files

//go:embed all:out
var content embed.FS

func main() {
	distFS, err := fs.Sub(content, "out")
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
