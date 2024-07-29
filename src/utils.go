package main 

import(
    "os"
    "path/filepath"
    "bytes"
    "log"
    "html/template"
    "github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
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

func MdToHtml(md []byte) []byte{
    extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
    p := parser.NewWithExtensions(extensions)
    doc := p.Parse(md)

    htmlFlags := html.CommonFlags | html.HrefTargetBlank
    opts := html.RendererOptions{Flags: htmlFlags}
    renderer := html.NewRenderer(opts)

    return markdown.Render(doc, renderer)
}

func RemovePostFromList(posts []Post, idToRemove string) []Post {
    for i, post := range posts {
        if post.Id.Hex() == idToRemove {
            return append(posts[:i], posts[i+1:]...)
        }
    }
    return posts
}
