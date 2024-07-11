package main

import(
    "net/http"
    "html/template"
    "log"
	"github.com/google/uuid"
)

func HandleEndpoints(){ 
	fs := http.FileServer(http.Dir("./dist"))
	http.Handle("/dist/", http.StripPrefix("/dist/", fs))
    http.HandleFunc("/", handleIndex)
    http.HandleFunc("/contact", handleContact)
    http.HandleFunc("/home", handleHome)
}

func handleIndex(w http.ResponseWriter, r *http.Request){
	tmpl, err := ParseTemplates()
	if err != nil {
		log.Printf("Error loading templates: %v\n", err)
		http.Error(w, "Error loading templates.", http.StatusInternalServerError)
		return
	}

    content, err := RenderTemplate(tmpl, "home", nil)
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
}

func handleContact(w http.ResponseWriter, r *http.Request){
		tmpl, err := ParseTemplates()
		if err != nil {
			log.Printf("Error loading templates: %v\n", err)
			http.Error(w, "Error loading templates.", http.StatusInternalServerError)
			return
		}

        content, err := RenderTemplate(tmpl, "contact", nil)
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
}

func handleHome(w http.ResponseWriter, r *http.Request){

		tmpl, err := ParseTemplates()
		if err != nil {
			log.Printf("Error loading templates: %v\n", err)
			http.Error(w, "Error loading templates.", http.StatusInternalServerError)
			return
		}

        content, err := RenderTemplate(tmpl, "home", nil)
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
}
