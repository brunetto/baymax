package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"

	"github.com/brunetto/baymax"
	"github.com/pkg/errors"
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
		b *baymax.Baymax
		err error
	)

	b, r, err = baymax.NewDefaultBaymaxWS("baymax.json", env)
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't create default Baymax WS"))
	}

	r.GET("/gui", b.MonitorGUI)
	r.GET("/json", b.MonitorJSON)

	// Start serving
	r.Run(":8081")
}
