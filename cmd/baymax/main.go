package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/brunetto/gin-logrus"
	"gopkg.in/gin-gonic/gin.v1"

	"github.com/brunetto/baymax"
	"github.com/brunetto/goutils/conf"
	"github.com/pkg/errors"
	"gitlab.com/brunetto/ritter"
)

var version string

func main() {

	// Print version
	var (
		v, vv bool
		env string
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
		r             *gin.Engine
		rotatedWriter *ritter.Writer
		err           error
		config        baymax.Conf
		b *baymax.Baymax
	)

	// Read conf
	err = conf.LoadJsonConf("baymax.json", &config)
	if err != nil {
		logrus.Fatalf(errors.Wrap(err, "can't load json config file baymax.json").Error())
	}

	// New writer with rotation
	rotatedWriter, err = ritter.NewRitterTime(config.LogFile)
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "can't create log file"))
	}

	// Tee to stderr
	rotatedWriter.TeeToStdErr = true

	// Create logger
	log := &logrus.Logger{
		Out:       rotatedWriter,
		Hooks: make(logrus.LevelHooks),
		Level: logrus.DebugLevel,
	}

	// Set text formatter options
	if env == "dev" {
		logFormatter := new(logrus.TextFormatter)
		logFormatter.FullTimestamp = true
		log.Formatter = new(logrus.TextFormatter)
	} else {
		log.Formatter = new(logrus.JSONFormatter)
	}

	// New locator
	b, err = baymax.NewBaymax()
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "can't create new locator"))
	}

	// New engine
	r = gin.New()

	// Set middlewere
	r.Use(ginlogrus.Logger(log), gin.Recovery())

	// Set routes
	// Service routes
	r.GET("/livecheck", func(c *gin.Context) { c.String(http.StatusOK, "%v", "OK") })
	r.GET("/favicon.ico", func(*gin.Context) { return })



	// Start serving
	r.Run(":8081")
}
