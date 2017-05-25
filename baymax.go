package baymax

import (
	"gopkg.in/gin-gonic/gin.v1"
	"io/ioutil"
	"strings"
	"net/http"
	"time"
	"github.com/pkg/errors"
	"github.com/brunetto/goutils/conf"
	"github.com/Sirupsen/logrus"
	"gitlab.com/brunetto/ritter"
	"github.com/brunetto/gin-logrus"
)

type Baymax struct {
	*http.Client
	Targets []Target
}

type Conf struct {
	Targets []Target
	LogLines int
	LogFile string
}

type Targets []Target

type Target struct {
	Name string `json:"name"`
	URL string `json:"url"`
	AlternativeURL string `json:"alternative_url"`
	LogLocation string `json:"log_location"`
	Status int `json:"status"`
	Message string `json:"message"`
	Logs string `json:"logs"`
}

func NewBaymax(t Targets) (*Baymax, error) {
	b := &Baymax{
		&http.Client{Timeout: 60 * time.Second},
		t,
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

	// New locator
	b, err = NewBaymax(config.Targets)
	if err != nil {
		return b, r, errors.Wrap(err, "can't create new locator")
	}

	// New engine
	r = gin.New()

	// Set middleware
	r.Use(ginlogrus.Logger(log), gin.Recovery())

	// Set routes
	// Service routes
	r.GET("/livecheck", func(c *gin.Context) { c.String(http.StatusOK, "%v", "OK") })
	r.GET("/favicon.ico", func(*gin.Context) { return })

	return b, r, err
}

func (b *Baymax) MonitorJSON (c *gin.Context) {
	b.CollectStatus()
	c.JSON(http.StatusOK, b.Targets)
	return
}

func (b *Baymax) MonitorGUI (c *gin.Context) {
	b.CollectStatus()
	c.HTML(http.StatusOK, "message", gin.H{
		"Title": "Status Monitor",
		"Targets": b.Targets,
	})
	return
}

func (b *Baymax) CollectStatus () {
	var (
		err error
	)
	for i, t := range b.Targets {
		err = b.ProbeTarget(t)
		if err != nil {
			t.Status = 0
			t.Message = err.Error()
		} else {
			t.Status = 1
			t.Message = "OK"
		}
		b.Targets[i] = t
	}
}

func (b *Baymax) CollectLogs () {}

func (b *Baymax) ProbeTarget(t Target) error {
	resp, err := b.Get(t.URL)
	if err != nil {
		return errors.Wrap(err, "can't reach the service")
	}

	if resp.StatusCode != 200 {
		return errors.Errorf("got %v (%v)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "unreadable response")
	}

	if strings.ToLower(string(body)) != "ok" {
		return errors.Errorf("service not ok: %v", string(body))
	}

	return nil
}
