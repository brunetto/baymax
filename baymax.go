package baymax

import (
	"gopkg.in/gin-gonic/gin.v1"
	"io/ioutil"
	"strings"
	"net/http"
	"github.com/pkg/errors"
	"fmt"
	"sort"
	"strconv"
	"os"
	"github.com/rogpeppe/rog-go/reverse"
	"github.com/brunetto/goutils/file"
	"time"
)

//go:generate rice embed-go

var Version string

type Baymax struct {
	*http.Client
	Targets []Target
	LogLines int
}

type Conf struct {
	Targets []Target `json:"targets"`
	LogLines int `json:"log_lines"`
	LogFile string `json:"log_file"`
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

func (b *Baymax) MonitorJSON (c *gin.Context) {
	b.CollectStatus()
	b.CollectLogs()
	c.JSON(http.StatusOK, b.Targets)
	return
}

func (b *Baymax) MonitorGUI (c *gin.Context) {
	ok, warns, errs := b.CollectStatus()
	b.CollectLogs()
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
		errs = append(errs, fmt.Errorf("service not ok: %v", string(body)))
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

func (b *Baymax) CollectLogs () {
	var (
		targets Targets
	)
	for _, t := range b.Targets {
		t = b.GetTargetLog(t)
		targets = append(targets, t)
	}
	sort.Sort(targets)
	b.Targets = targets
	return
}

func (b *Baymax) GetTargetLog (t Target) Target {
	if t.LogLocation == "" {
		t.Logs = errors.New("no log specified").Error()
		return t
	}
	logFile := strings.Replace(t.LogLocation, "{{date}}", time.Now().Format("2006-01-02"), 1)

	if !file.Exists(logFile) {
		t.Logs = "log file doesn't exists"
		return t
	}

	logFileObj, err := os.Open(logFile)
	if err != nil {
		t.Logs = errors.Wrap(err, "can't open log file").Error()
		return t
	}
	defer logFileObj.Close()

	logs := []string{}
	count := 0
	scanner := reverse.NewScanner(logFileObj)
	fmt.Println(logFile, b.LogLines)
	for scanner.Scan() {
		if count >= b.LogLines {
			break
		}
		line := scanner.Text()
		fmt.Println("line ", line)
		logs = append(logs, line)
		count++
	}
	// Reverse the order
	for i:=1; i<=len(logs); i++ {
		t.Logs += logs[len(logs)-i]
	}

	return t
}
