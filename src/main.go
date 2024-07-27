package main

import (
	"html/template"
	"log"
	"net/http"
)

type PageData struct {
	Content template.HTML
	Version string
}

func main() {
    ConnectMongoDB()
    HandleEndpoints()
	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", fs))
    HandleAdminEndpoints()
	log.Println("Server started at :8880")
	if err := http.ListenAndServe("[::]:8880", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
