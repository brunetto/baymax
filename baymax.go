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
	"github.com/GeertJohan/go.rice"
	"html/template"
	"fmt"
	"sort"
	"strconv"
	"gitlab.com/brunetto/hang"
)

//go:generate rice embed-go

var Version string

type Baymax struct {
	*http.Client
	Targets []Target
}

type Conf struct {
	Targets []Target
	LogLines int
	LogFile string
}

type errorSlice []error
func (es errorSlice) Error() string {
	switch len(es) {
	case 0:
		return ""
	case 1:
		return es[0].Error()
	default:
		err := errors.Errorf("[0] %v", es[0].Error())
		for i, e := range es[1:] {
			err = errors.Wrap(err, "[" + strconv.Itoa(i+1) + "] " + e.Error()+"\n")
		}
		return err.Error()
	}
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
// Sort targets on name
func (t Targets) Len() int           { return len(t) }
func (t Targets) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t Targets) Less(i, j int) bool { return t[i].Name < t[j].Name }

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

	hang.LogStartAndStop("baymax", log)

	// New locator
	b, err = NewBaymax(config.Targets)
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

func (b *Baymax) MonitorJSON (c *gin.Context) {
	b.CollectStatus()
	c.JSON(http.StatusOK, b.Targets)
	return
}

func (b *Baymax) MonitorGUI (c *gin.Context) {
	ok, warns, errs := b.CollectStatus()
	c.HTML(http.StatusOK, "message", gin.H{
		"Title": "Status Monitor",
		"Targets": b.Targets,
		"Ok": ok, "Warnings": warns, "Errors": errs,
	})
	return
}

func (b *Baymax) CollectStatus () (ok, warns, errs int) {
	var (
		targets Targets
	)
	for _, t := range b.Targets {
		t = b.ProbeTarget(t)
		targets = append(targets, t)
		switch t.Status {
		case 0:
			errs += 1
		case 1:
			ok +=1
		case 2:
			warns +=1
		}
	}
	sort.Sort(targets)
	b.Targets = targets
	return ok, warns, errs
}

func (b *Baymax) CollectLogs () {}

func (b *Baymax) ProbeTarget(t Target) Target {
	var (
		resp *http.Response
		err error
		errs errorSlice
	)
	resp, err = b.Get(t.URL)
	if err != nil {
		errs = append(errs, errors.Wrap(err, "can't reach service on main URL"))
		if t.AlternativeURL != "" {
			errs = append(errs, errors.Errorf("probing alternative URL: %v", t.AlternativeURL))
			resp, err = b.Get(t.AlternativeURL)
			if err != nil {
				errs = append(errs, errors.Wrap(err, "can't reach service on alternative URL"))
				t.Status = 0
				t.Message = errs.Error()
				return t
			}
		} else {
			errs = append(errs, errors.New("no alternative URL provided"))
			t.Status = 0
			t.Message = errs.Error()
			return t
		}
	}

	if resp.StatusCode != 200 {
		t.Status = 2
		errs =  append(errs, errors.Errorf("got %v (%v)", resp.StatusCode, http.StatusText(resp.StatusCode)))
		t.Message = errs.Error()
		return t
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Status = 2
		errs = append(errs, errors.Wrap(errs, "unreadable response"))
		t.Message = errs.Error()
		return t
	}

	if strings.ToLower(string(body)) != "ok" {
		t.Status = 2
		errs = append(errs, errors.Wrap(err, fmt.Sprintf("service not ok: %v", string(body))))
		t.Message = errs.Error()
		return t
	}

	if len(errs) != 0 {
		t.Status = 2
		t.Message = errs.Error()
	} else {
		t.Status = 1
		t.Message = "OK"
	}
	return t
}
