// httpTest.go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"

	// "stringUtils"
	"strings"
	"time"
)

type TestUnit struct {
	UserId  string            `json:"-"`
	Enabled bool              `json:"enabled"`
	reqNum  int               `json:"reqNum"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Cookies map[string]string `json:"coockies"`
	Body    map[string]string `json:"body"`
}
type LoginConfig struct {
	Request TestUnit `json:"request"`
}

type ConfigObj struct {
	Protocol      string     `json:"protocol"`
	Host          string     `json:"host"`
	Port          string     `json:"port"`
	AccountPrefix string     `json:"accountPrefix"`
	AccountFrom   int        `json:"accountFrom"`
	AccountTo     int        `json:"accountTo"`
	Passwd        string     `json:"passwd"`
	JoinInterval  int64      `json:"joinInterval"`
	NextDelay     int64      `json:nextDelay`
	PreRequests   []TestUnit `json:"preRequests"`
	RandRequests  []TestUnit `json:"randRequests"`
}

type SequeenRequests struct {
	Reqs      []*TestUnit
	SendIndex int
	NextDelay int64
}

var (
	preUrl = ""
)

func main() {
	fmt.Println("cpu num:", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	configFile := "config_http_demo.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	protocol := "https"
	if len(os.Args) > 2 {
		configFile = os.Args[2]
	}
	switch protocol {
	case "http":
	case "https":
		confObj := parseHttpConfig("configs/" + configFile)
		// reqs := initRequests(confObj)
		StartHttpTest(confObj)
	case "tcp":

	}

	// exit event
	fmt.Println("[ctrl+c to exit]")
	end := make(chan os.Signal, 2)
	signal.Notify(end, os.Interrupt, os.Kill)
	<-end
}

func parseHttpConfig(fileName string) *ConfigObj {
	fmt.Println("----init http config:", fileName)
	conf, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	confObj := &ConfigObj{}
	noRemarkCont := string(conf)
	noRemarkCont = ReplaceComment(noRemarkCont)
	err = json.Unmarshal([]byte(noRemarkCont), confObj)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return confObj
}

func buildRequest(testUnit *TestUnit) *http.Request {
	sendData := url.Values{}

	for k, v := range testUnit.Body {
		sendData.Add(k, v)
	}
	req, err := http.NewRequest(testUnit.Method, preUrl+testUnit.Path, strings.NewReader(sendData.Encode()))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return nil
	}
	//header
	for k, v := range testUnit.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("uid", testUnit.UserId)

	//cookie
	for k, v := range testUnit.Cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}

	return req
}

func StartHttpTest(confObj *ConfigObj) {
	fmt.Println("----StartHttpTest----")
	preUrl = confObj.Protocol + "://" + confObj.Host + confObj.Port
	interval := confObj.JoinInterval
	lastTime := time.Now().UnixNano() / int64(time.Millisecond)
	totalUser := confObj.AccountTo - confObj.AccountFrom
	fmt.Println("totalUser:", totalUser)
	loginIndex := 0
	for {
		nowTime := time.Now().UnixNano() / int64(time.Millisecond)
		if nowTime-lastTime > interval {
			lastTime = nowTime
			currLoginIndex := loginIndex
			if loginIndex >= totalUser {
				fmt.Println("==== all users have start session ! =====")
				break
			}
			fmt.Println("----startSession:", currLoginIndex)
			go startSession(currLoginIndex, confObj)
			loginIndex++
		}
	}
}

func startSession(currLoginIndex int, confObj *ConfigObj) {
	userId := confObj.AccountPrefix + strconv.Itoa(confObj.AccountFrom+currLoginIndex)
	// var token string
	// passwd := confObj.Passwd//t
	// fmt.Println("userId:", userId)

	//优先请求
	preRequests := []*TestUnit{}
	for _, testUnit := range confObj.PreRequests {
		reqConf := testUnit
		reqConf.UserId = userId
		if !reqConf.Enabled {
			continue
		}
		preRequests = append(preRequests, &reqConf)
	}
	// fmt.Println("----preRequests len:", len(preRequests))
	if len(preRequests) > 0 {
		sequeenReqs := &SequeenRequests{
			Reqs:      preRequests,
			NextDelay: confObj.NextDelay,
		}
		sendSequenRequests(sequeenReqs)
	}

	//循环请求
	randRequests := []*TestUnit{}
	for _, testUnit := range confObj.RandRequests {
		reqConf := testUnit
		reqConf.UserId = userId
		if !reqConf.Enabled {
			continue
		}
		randRequests = append(randRequests, &reqConf)
	}
	// fmt.Println("----randRequests len:", len(randRequests))
	if len(randRequests) > 0 {
		sendRandRequests(confObj.NextDelay, randRequests)
	}
}

//一次性顺序请求
func sendSequenRequests(sequeenReq *SequeenRequests) {
	if sequeenReq.SendIndex >= len(sequeenReq.Reqs) {
		// fmt.Println("----sendSequenRequests finish-----")
		return
	}
	sendRequest(sequeenReq.Reqs[sequeenReq.SendIndex])
	sequeenReq.SendIndex++
	time.Sleep(time.Duration(sequeenReq.NextDelay) * time.Second)
	sendSequenRequests(sequeenReq)
}

//重复性随机请求
func sendRandRequests(nextDelay int64, randRequests []*TestUnit) {
	requestIndex := rand.Intn(len(randRequests))
	sendRequest(randRequests[requestIndex])
	time.Sleep(time.Duration(nextDelay) * time.Second)
	sendRandRequests(nextDelay, randRequests)
}

func sendRequest(reqConf *TestUnit) []byte {
	req := buildRequest(reqConf)
	// fmt.Println("----sendRequest:", req.URL)
	rsp, err := (&http.Client{}).Do(req)
	if err != nil {
		fmt.Println(err)
		return []byte("{}")
	}
	var respData []byte
	if respData, err = ioutil.ReadAll(rsp.Body); err != nil {
		return []byte("{}")
	}
	defer rsp.Body.Close()
	respStr := string(respData)
	if strings.Index(respStr, "{\"code\":0,") != 0 {
		fmt.Println("rsp:", respStr)
	}
	return respData
}

//-----------------
//删除代码中的//和/**/注释
func ReplaceComment(noRemarkCont string) string {
	lineRegPatten := `\/\/[^\n]*`
	blockRegPatten := `\/\*.*?\*\/`

	lineReg, _ := regexp.Compile(lineRegPatten)
	blockReg, _ := regexp.Compile(blockRegPatten)

	noRemarkCont = lineReg.ReplaceAllString(noRemarkCont, "")
	noRemarkCont = blockReg.ReplaceAllString(noRemarkCont, "")
	return noRemarkCont
}
