package main

import (
	"log"
	"net/http"
	"github.com/vinny-pereira/personal-blog/api"
	"github.com/vinny-pereira/personal-blog/internal/repository"
)


func main() {
    repository.ConnectMongoDB()
    api.HandleEndpoints()
	fs := http.FileServer(http.Dir("./web/wwwroot/uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", fs))
    api.HandleAdminEndpoints()
	log.Println("Server started at :8880")
	if err := http.ListenAndServe("[::]:8880", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
