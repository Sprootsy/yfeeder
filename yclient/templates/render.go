package templates

import (
	"bufio"
	"context"
	"embed"
	"html/template"
	"log"
	"os"
	"path"
	"yclient/model"
)

//go:embed html/*
var templatesFS embed.FS

// Templates parsed templates
var Templates *template.Template

var Names = []string{"html/articles.html"}

func init() {
	log.Println("Loading templates")
	tmpls, errParseFS := template.ParseFS(templatesFS, Names...)
	if errParseFS != nil {
		log.Fatalln("Cannot parse templates. Error:", errParseFS)
	}
	Templates = tmpls
	log.Println("Finished loading templates")
}

func Render(ctx context.Context, outDir string, articles []model.Article) {
	log.Println("Rendering templates in", outDir)
	for _, n := range Names {
		tn := path.Base(n)
		t := Templates.Lookup(tn)
		if t == nil {
			log.Println("Template not found:", tn)
		} else {
			log.Println("Executing template:", tn)
			f, errOpen := os.OpenFile(path.Join(outDir, tn), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0754)
			if errOpen != nil {
				log.Println("Cannot open file for writing. Error:", errOpen)
				continue
			}
			w := bufio.NewWriter(f)
			if errTmpl := t.Execute(w, articles); errTmpl != nil {
				log.Println("Cannot execute template. Error:", errTmpl)
			}
			w.Flush()
			f.Close()
		}
	}
	log.Println("Done rendering templates")
}
