package templates

import (
	"bufio"
	"embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"yclient/model"
)

//go:embed html/*
var templatesFS embed.FS

var (
	templates *template.Template
	// Files list of files that are loaded as templates
	Files     = []string{"html/articles.html", "html/_article.ID_.html"}
)

type Renderable interface {
	Render() error
}

type Rendering struct {
	Name     string
	Path     string
	Template *template.Template
}

type ArticlesRendering struct {
	Rendering
	Data []model.Article
}

func (t ArticlesRendering) Render() error {
	fileName := path.Join(t.Path, "articles.html")	
	return Execute(t.Template, fileName, t.Data)
}

type Summary struct {
	Title	 string
	Summary  template.HTML
}

type SummariesRendering struct {
	Rendering
	Data	model.Article
}

func (r SummariesRendering) Render() error {
	fileName := path.Join(r.Path, fmt.Sprintf("%s_summary.html", r.Data.ID))
	return Execute(r.Template, fileName, Summary{
		Title: r.Data.Title,
		Summary: template.HTML(r.Data.Summary),
	})
}

func init() {
	log.Println("Loading templates")

	tmpls, errParseFS := template.ParseFS(templatesFS, Files...)
	if errParseFS != nil {
		log.Fatalln("Cannot parse templates. Error:", errParseFS)
	}
	templates = tmpls
	log.Println("Finished loading templates, found:", len(tmpls.Templates()))
}

// GetTemplate returns the template associated with the given file
func GetTemplate(file string) *template.Template {
	return templates.Lookup(path.Base(file))
}

// Execute executes the given template with the given `data`.
// The output is written in `path`. Any error is returned
func Execute(t *template.Template, path string, data any) error {
	if t == nil {
		return fmt.Errorf("Cannot execute template: nil pointer.")
	}
	log.Println("Executing", t.Name(), " in:", path)
	fd, errOpen := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0754)
	if errOpen != nil {
		return errOpen
	}

	w := bufio.NewWriter(fd)
	if errEx := t.Execute(w, data); errEx != nil {
		defer fd.Close()
		return errEx
	}
	if errFlush := w.Flush(); errFlush != nil {
		defer fd.Close()
		return errFlush
	}
	if errClose := fd.Close(); errClose != nil {
		return errClose
	}

	log.Println("Done executing", t.Name())
	return nil
}
