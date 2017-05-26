package baymax

import (
	"github.com/Sirupsen/logrus"
	"gitlab.com/brunetto/hang"
	"github.com/GeertJohan/go.rice"
	"time"
	"gitlab.com/brunetto/ritter"
	"github.com/brunetto/goutils/conf"
	"html/template"
	"github.com/pkg/errors"
	"gopkg.in/gin-gonic/gin.v1"
	"github.com/brunetto/gin-logrus"
	"net/http"
)

func NewBaymax(c        Conf) (*Baymax, error) {
	b := &Baymax{
		Client: &http.Client{Timeout: 60 * time.Second},
		Targets: c.Targets,
		LogLines: c.LogLines,
	}
	return b, nil
}

func NewDefaultBaymaxWS(configFile string, env string) (*Baymax, *gin.Engine, error) {
	var (
		err error
		config        Conf
		rotatedWriter *ritter.Writer
		b *Baymax
		r *gin.Engine
	)
	// Read conf
	err = conf.LoadJsonConf(configFile, &config)
	if err != nil {
		return b, r, errors.Wrap(err, "can't load json config file baymax.json")
	}

	// New writer with rotation
	rotatedWriter, err = ritter.NewRitterTime(config.LogFile)
	if err != nil {
		return b, r, errors.Wrap(err, "can't create log file")
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

	hang.LogStartAndStop("baymax", log)

	// New locator
	b, err = NewBaymax(config)
	if err != nil {
		return b, r, errors.Wrap(err, "can't create new locator")
	}

	// New engine
	r = gin.New()

	// Load templates
	templateBox, err := rice.FindBox("assets/templates")
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

	// Set middleware
	r.Use(ginlogrus.Logger(log), gin.Recovery())

	// Set routes
	// Service routes
	r.GET("/livecheck", func(c *gin.Context) { c.String(http.StatusOK, "%v", "OK") })
	r.GET("/favicon.ico", func(*gin.Context) { return })

	return b, r, err
}

