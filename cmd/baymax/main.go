package main

import (
	"flag"
	"fmt"
	"gopkg.in/gin-gonic/gin.v1"

	"github.com/GeertJohan/go.rice"
	"github.com/brunetto/baymax"
	"github.com/pkg/errors"
	"html/template"
	"log"
)

var version string

func main() {

	// Print version
	var (
		v, vv bool
		env   string
	)
	flag.BoolVar(&v, "v", false, "Print version and exit")
	flag.BoolVar(&vv, "version", false, "Print version and exit")
	flag.StringVar(&env, "e", "prod", "Environment [dev, prod]")

	flag.Parse()

	if v || vv {
		fmt.Println(version)
		return
	}

	var (
		r *gin.Engine

		err error

		b *baymax.Baymax
	)

	b, r, err = baymax.NewDefaultBaymaxWS("baymax.json", env)
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't create default Baymax WS"))
	}

	templateBox, err := rice.FindBox("../../assets/templates")
	if err != nil {
		log.Fatal(err)
	}
	// get file contents as string
	templateString, err := templateBox.String("index.html.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	// parse and execute the template
	tmplMessage, err := template.New("message").Parse(templateString)
	if err != nil {
		log.Fatal(err)
	}
	r.SetHTMLTemplate(tmplMessage)

	r.GET("/gui", b.MonitorGUI)

	r.GET("/json", b.MonitorJSON)

	// Start serving
	r.Run(":8081")
}
