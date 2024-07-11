package main 

import(
    "os"
    "path/filepath"
    "bytes"
    "log"
    "html/template"
)

func RenderTemplate(tmpl *template.Template, name string, data interface{}) (string, error) {
    var buf bytes.Buffer
    err := tmpl.ExecuteTemplate(&buf, name, data)
    if err != nil {
        log.Printf("Error rendering template %s: %v\n", name, err)
    }
    return buf.String(), nil
}


func ParseTemplates() (*template.Template, error) {
	tmpl := template.New("")
	err := filepath.Walk("./views", func(path string, info os.FileInfo, err error) error {
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
