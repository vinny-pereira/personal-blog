package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"github.com/google/uuid"
    "bytes"
)

type PageData struct {
	Content template.HTML
	Version string
}

func main() {
	fs := http.FileServer(http.Dir("./dist"))
	http.Handle("/dist/", http.StripPrefix("/dist/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := parseTemplates("./views")
		if err != nil {
			log.Printf("Error loading templates: %v\n", err)
			http.Error(w, "Error loading templates.", http.StatusInternalServerError)
			return
		}

        content, err := renderTemplate(tmpl, "home", nil)
        if err != nil {
            log.Printf("Error rendering home template: %v\n", err)
            http.Error(w, "Error rendering template.", http.StatusInternalServerError)
            return
        }

		version := uuid.New()
		data := PageData{
			Content: template.HTML(content),
			Version: version.String(),
		}

		if err := tmpl.ExecuteTemplate(w, "index", data); err != nil {
			log.Println(err)
			http.Error(w, "Error executing template.", http.StatusInternalServerError)
		}
	})

    http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request){

		tmpl, err := parseTemplates("./views")
		if err != nil {
			log.Printf("Error loading templates: %v\n", err)
			http.Error(w, "Error loading templates.", http.StatusInternalServerError)
			return
		}

        content, err := renderTemplate(tmpl, "contact", nil)
        if err != nil {
            log.Printf("Error rendering contact template: %v\n", err)
            http.Error(w, "Error rendering template.", http.StatusInternalServerError)
            return
        }

		version := uuid.New()
		data := PageData{
			Content: template.HTML(content),
			Version: version.String(),
		}

		if err := tmpl.ExecuteTemplate(w, "contact", data); err != nil {
			log.Println(err)
			http.Error(w, "Error executing template.", http.StatusInternalServerError)
		}
    })

    http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request){

		tmpl, err := parseTemplates("./views")
		if err != nil {
			log.Printf("Error loading templates: %v\n", err)
			http.Error(w, "Error loading templates.", http.StatusInternalServerError)
			return
		}

        content, err := renderTemplate(tmpl, "home", nil)
        if err != nil {
            log.Printf("Error rendering contact template: %v\n", err)
            http.Error(w, "Error rendering template.", http.StatusInternalServerError)
            return
        }

		version := uuid.New()
		data := PageData{
			Content: template.HTML(content),
			Version: version.String(),
		}

		if err := tmpl.ExecuteTemplate(w, "home", data); err != nil {
			log.Println(err)
			http.Error(w, "Error executing template.", http.StatusInternalServerError)
		}
    })

	log.Println("Server started at :8880")
	if err := http.ListenAndServe("[::]:8880", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}

func renderTemplate(tmpl *template.Template, name string, data interface{}) (string, error) {
    var buf bytes.Buffer
    err := tmpl.ExecuteTemplate(&buf, name, data)
    if err != nil {
        log.Printf("Error rendering template %s: %v\n", name, err)
    }
    return buf.String(), nil
}

func parseTemplates(dir string) (*template.Template, error) {
	tmpl := template.New("")
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".html" {
			_, err := tmpl.ParseFiles(path)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

