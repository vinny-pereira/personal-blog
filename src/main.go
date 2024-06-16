package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type PageData struct{
    Name string
}

func main(){
    var viewsDir = "../views";
    var template_paths []string;
    templates := make(map[string]*template.Template);

    err := filepath.Walk(viewsDir, func(path string, info os.FileInfo, err error) error{
        if err != nil{
            return err;
        }

        if !info.IsDir(){
            relPath, err := filepath.Rel(viewsDir, path);

            if err != nil{
                return err;
            }

            template_paths = append(template_paths, relPath);
        }

        return nil
    });

    if err != nil{
        fmt.Printf("Error walking the views directory: %v\n", err);
        return;
    }

    for _, v := range template_paths{
        tmpl, err := template.ParseFiles(v);

        if err != nil{
            continue;
        }

        templates[tmpl.Name()] = tmpl;
    }

    fs := http.FileServer(http.Dir("../dist"));

    http.Handle("/dist", http.StripPrefix("dist", fs));
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
        tmpl := templates["index"];
        data := PageData{
            Name: "Blog",
        }

        if err := tmpl.ExecuteTemplate(w, "index", data); err != nil{
            http.Error(w, "Error executing template.", http.StatusInternalServerError);
        }
    });

    log.Println("Server started at :8880");
    if err := http.ListenAndServe("[::]:8880", nil); err != nil{
        log.Fatalf("Could not start server: %s\n", err);
    }
}

