package baymax

import (
	"gopkg.in/gin-gonic/gin.v1"
	"io/ioutil"
	"strings"
	"net/http"
	"time"
	"github.com/pkg/errors"
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
