package requests

import (
	"fmt"
	//	"context"
	"encoding/base64"
	"github.com/raff/godet"
	"github.com/rezen/retirejs"
	"math/rand"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// https://github.com/tebeka/selenium

var versionsScript = `
return function() {
  var versions = Object.keys(window)
    .map(k => [
      k,
      Object.keys(window[k] || "")
       .filter(l => l.toLocaleLowerCase().indexOf("version") !== -1)
       .map(l => ["string", "number"].indexOf(typeof window[k][l]) !== -1 ? ("" + window[k][l]) : "")
       .pop()
    ])
    .filter(t => !!t[1])
    .map(v => v.map(z => "" + z));
    versions.push(["algolia", "__algolia" in window ? __algolia.algoliasearch.version : null]);
[PLACEHOLDER]
   return versions;
}();`

func scriptForVersions() string {
	repo := retirejs.GetRepository()
	funcs := repo.FuncsForDomVersions()
	script := ""
	for _, version := range funcs {
		script += fmt.Sprintf("    try { versions.push(['%versionv', %v]); } catch (e) {};\n", version[0], version[1])
	}
	return strings.Replace(versionsScript, "[PLACEHOLDER]", script, -1)
}

type BrowserConfig struct {
	Timeout           int
	Port              int
	Block             []string
	UseAuthentication bool
	Username          string
	Password          string
	Proxy             string
	Scripts           map[string]string // Scripts execute
	RunBeforeGet      func(*godet.RemoteDebugger)
	RunAfterGet       func(*godet.RemoteDebugger)
}

func DefaultBrowserConfig() BrowserConfig {
	return BrowserConfig{
		Timeout:           3,
		Port:              9200,
		Block:             []string{"*.svg", "*.jpg", "*.png", "*.gif", "*.css", "*.woff2", "*.woff"},
		UseAuthentication: false,
		// Proxy: "127.0.0.1:8080",
	}
}

type BrowserTransaction struct {
	Target       string
	TargetIp     string
	Code         int
	Protocol     string
	ResponseBody string
	BodyHash     string
	ResHeaders   map[string]string
	Security     map[string]string
	MimeType     string
}

func (t BrowserTransaction) Body() string {
	return t.ResponseBody
}

func (t BrowserTransaction) Url() *url.URL {
	parsed, _ := url.Parse(t.Target)
	return parsed
}

func (t BrowserTransaction) Header(key string) string {
	if val, ok := t.ResHeaders[key]; ok {
		return val
	}

	return ""
}

func (t BrowserTransaction) Headers() map[string]string {
	headers := map[string]string{}
	if t.ResHeaders == nil {
		return headers
	}

	for key, value := range t.ResHeaders {
		headers[strings.ToLower(key)] = value
	}
	return headers
}

func (t BrowserTransaction) StatusCode() int {
	return t.Code
}

func TransactionFromResponse(res map[string]interface{}, body string) BrowserTransaction {
	path := res["url"].(string)
	data := infermap(res)

	headers := data.MapStringString("headers")
	security := data.MapStringString("securityDetails")
	ip := data.String("remoteIPAddress")

	return BrowserTransaction{
		Target:   path,
		Protocol: data.String("protocol"),
		TargetIp: ip,
		Code:     int(data.Float("status")),
		// ResponseBody: body,
		ResHeaders: headers,
		Security:   security,
		MimeType:   data.String("mimeType"),
	}
}

func executeScripts(browser *godet.RemoteDebugger) map[string]interface{} {
	scripts := map[string]string{
		"versions": scriptForVersions(),
		"dom_html": "return (() => document.documentElement.outerHTML)()",
	}

	results := map[string]interface{}{}
	for key, script := range scripts {
		time.Sleep(time.Millisecond * 100)
		res, err := browser.EvaluateWrap(script)
		results[key] = res
		fmt.Println("script-err", err)
	}
	return results
}

type Page struct {
	Tab             *godet.Tab
	FrameId         string
	TargetId        string
	UserAgent       string
	Goto            string
	Durtion         time.Duration
	ResponseCounter int
	RunAfterLoad    string
	RunBeforeClose  string
	Browser         *godet.RemoteDebugger
	NetworkIdle     chan bool
}

func (p *Page) SetRequestHeaders(map[string]string) {
	/*
		if config.UseAuthentication {
			// @todo move to user?
			auth := config.Username + ":" + config.Password
			encoded :=  base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Add("Authorization","Basic " + encoded)
		}

		for header, value := range config.Headers {
			req.Header.Add(header, value)
		}
	*/
}

// @todo intercept file
func (p *Page) onResponse(params godet.Params) {
	go func() {
		body := ""
		// @todo if mode is quick
		response := params["response"].(map[string]interface{})

		res, err := p.Browser.SendRequest("Network.getResponseBody", godet.Params{
			"requestId": params["requestId"],
		})

		if err != nil {
			// @todo log
		}

		if res["body"] != nil {
			body = res["body"].(string)
		}

		if res["base64Encoded"] != nil && res["base64Encoded"].(bool) {
			text, _ := base64.StdEncoding.DecodeString(res["body"].(string))
			body = string(text[:])
		}

		TransactionFromResponse(response, body)

		// fmt.Println(transaction)

		// r.Waiter.Done()
		// r.Responses<-transaction

	}()
}

func (p *Page) Get(path string) {
	// @todo ensure tab is in focus
	p.Browser.ActivateTab(p.Tab)
	p.Browser.Navigate(path)
	fmt.Println("WENT THERE?")
}

func (p *Page) SetupEvents(browser *godet.RemoteDebugger) {
	//	blocked := []string{}

	/*
		browser.CallbackEvent("Runtime.executionContextCreated", func(params godet.Params) {
			fmt.Println("CTX-CREATE", params)
		})*/

	browser.CallbackEvent("Page.lifecycleEvent", func(params godet.Params) {
		go func() {
			event := params["name"].(string)
			frameId := params["frameId"].(string)
			if frameId == p.FrameId {
				// fmt.Println("EVENT", event, params)
				if event == "networkIdle" {
					// fmt.Println("SET-IDLE")
					p.NetworkIdle <- true
				}
			}
		}()
	})

	//browser.CallbackEvent("Network.responseReceived", p.onResponse)
	/*
		browser.CallbackEvent("Network.requestWillBeSent", func(params godet.Params) {
			p.ResponseCounter++
			// r.Waiter.Add(1)
			/*
			go func() {
				req := params["request"].(map[string]interface{})
				r.Requested<-req["url"].(string)
			}()

		})
	*/

	go func() {
		browser.SendRequest("Page.setBypassCSP", godet.Params{"enable": true})
		browser.SendRequest("Target.setAutoAttach", godet.Params{"autoAttach": true, "waitForDebuggerOnStart": false})
		browser.SendRequest("Page.setLifecycleEventsEnabled", godet.Params{"enabled": true})
		/*
			browser.SendRequest("Network.enable", godet.Params{})
			browser.SendRequest("Runtime.enable", godet.Params{})
			browser.SendRequest("Security.enable", godet.Params{})
			browser.SendRequest("Performance.enable", godet.Params{})
			browser.SendRequest("Log.enable", godet.Params{})
		*/
		//browser.SetBlockedURLs(blocked...)
	}()

	time.Sleep(time.Millisecond * 100)
}

func CreatePage(browser *godet.RemoteDebugger) *Page {
	tab, err := browser.NewTab("about:blank")

	if err != nil {
		panic(err)
	}

	page := &Page{
		Tab:         tab,
		FrameId:     tab.ID,
		Browser:     browser,
		NetworkIdle: make(chan bool),
	}

	// events needs to be associated to current tab (enable AFTER NewTab)
	// https://github.com/raff/godet/issues/32#issuecomment-368785126

	page.SetupEvents(browser)
	return page
}

type BrowserRequestor struct {
	Target          string
	Base            *url.URL
	History         map[string]*Transaction
	ByHash          map[string]*Transaction
	Config          BrowserConfig
	Browser         *godet.RemoteDebugger
	Tabs            []*godet.Tab
	Args            []string
	Port            int
	Debug           bool
	Cmd             *exec.Cmd
	ResponseCounter int
	Responses       chan BrowserTransaction
	Requested       chan string
	Waiter          sync.WaitGroup
	ContextId       string
	SessionId       string
}

func (r *BrowserRequestor) Get(path string) {
	// @todo url validate
	start := time.Now()

	// @todo add tab for cleanup
	page := CreatePage(r.Browser)

	// defer r.Browser.CloseTab(page.Tab)

	page.Get(path)
	fmt.Println("WAITING????")

	select {
	// @todo configure coninute after
	case <-time.After(time.Millisecond * 4000):
		fmt.Println("ONWARD, NO WAITING")
	case <-page.NetworkIdle:
		fmt.Println("ALLL FINISHED?!")
	}

	time.Sleep(time.Millisecond * 200)

	// @todo screenshot
	// err = r.Browser.SaveScreenshot("screenshot.png", 0644, 0, true)
	// raw, err := remote.CaptureScreenshot("png", 0, true)

	// @todo pass in config
	results := executeScripts(r.Browser)

	fmt.Sprintf("", results)
	fmt.Println("DONE")
	duration := time.Since(start)
	fmt.Println("Browser", duration)
}

func (r *BrowserRequestor) SetRequestHeaders(config *BrowserConfig) {
	/*
		if config.UseAuthentication {
			// @todo move to user?
			auth := config.Username + ":" + config.Password
			encoded :=  base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Add("Authorization","Basic " + encoded)
		}

		for header, value := range config.Headers {
			req.Header.Add(header, value)
		}
	*/
}

// func (r *BrowserRequestor)

// @todo regname setup
func (r *BrowserRequestor) Setup() {
	var remote *godet.RemoteDebugger
	var err error
	cmd := exec.Command("chromium-browser", r.Args...)
	err = cmd.Start()
	host := fmt.Sprintf("localhost:%v", r.Port)

	for i := 0; i < 12; i++ {
		if i > 0 {
			time.Sleep(300 * time.Millisecond)
		}

		remote, err = godet.Connect(host, r.Debug)
		if err == nil {
			break
		}
	}

	remote.SendRequest("Network.enable", godet.Params{})
	remote.SendRequest("Network.enable", godet.Params{})
	remote.SendRequest("Runtime.enable", godet.Params{})
	remote.SendRequest("Security.enable", godet.Params{})
	remote.SendRequest("Performance.enable", godet.Params{})
	remote.SendRequest("Log.enable", godet.Params{})

	r.Cmd = cmd
	r.Browser = remote

	r.Browser.CallbackEvent("Target.attachedToTarget", func(params godet.Params) {
		r.SessionId = params["sessionId"].(string)
		targetInfo := params["targetInfo"].(map[string]interface{})
		r.ContextId = targetInfo["browserContextId"].(string)
	})

	r.Browser.CallbackEvent("Runtime.executionContextCreated", func(params godet.Params) {
		fmt.Println("CONTEXT CREATED")
		fmt.Println(params)
	})

	r.Browser.CallbackEvent("Network.responseReceived", func(params godet.Params) {
		go func() {
			body := ""
			// @todo if mode is quick
			response := params["response"].(map[string]interface{})

			res, err := r.Browser.SendRequest("Network.getResponseBody", godet.Params{
				"requestId": params["requestId"],
			})

			if err != nil {
				// @todo log
			}

			if res["body"] != nil {
				body = res["body"].(string)
			}

			if res["base64Encoded"] != nil && res["base64Encoded"].(bool) {
				text, _ := base64.StdEncoding.DecodeString(res["body"].(string))
				body = string(text[:])
			}

			transaction := TransactionFromResponse(response, body)
			fmt.Println(transaction)

			// r.Waiter.Done()
			// r.Responses<-transaction

		}()
	})

	/*
		r.Browser.CallbackEvent("Network.responseReceived", r.onResponse)
		r.Browser.CallbackEvent("Network.requestWillBeSent", func(params godet.Params) {
			r.ResponseCounter++
			// r.Waiter.Add(1)
			go func() {
				req := params["request"].(map[string]interface{})
				r.Requested<-req["url"].(string)
			}()
		})
	*/
}

func (r *BrowserRequestor) Cleanup() {
	time.Sleep(time.Millisecond * 100)
	if err := r.Cmd.Process.Kill(); err != nil {
		fmt.Println("failed to kill process: ", err)
	}
}

func CreateBrowser() *BrowserRequestor {
	transactions := map[string]*Transaction{}
	config := DefaultBrowserConfig()
	fmt.Println()

	args := []string{
		// "--headless",
		"--user-data-dir=/tmp/chrome-" + fmt.Sprintf("%v", rand.Int()),
		"--no-first-run",
		"--remote-debugging-port=9222",
		"--hide-scrollbars",
		"--disable-background-networking",
		"--disable-background-timer-throttling",
		"--disable-breakpad",
		"--disable-client-side-phishing-detection",
		"--disable-default-apps",
		"--disable-dev-shm-usage",
		"--disable-extensions",
		"--disable-features=site-per-process",
		"--disable-hang-monitor",
		"--disable-popup-blocking",
		"--disable-prompt-on-repost",
		"--disable-sync",
		"--disable-translate",
		"--metrics-recording-only",
		"--safebrowsing-disable-auto-update",
		"--enable-automation",
		"--password-store=basic",
		"--use-mock-keychain",
		"--mute-audio",
	}

	if len(config.Proxy) > 0 {
		args = append(args, "--proxy-server="+config.Proxy)
		args = append(args, "--ignore-certificate-errors")
		args = append(args, "--proxy-bypass-list='*.gstatic.com;'")
		// @todo log ...
	}

	return &BrowserRequestor{
		Target:          "",
		Port:            9222,
		Args:            args,
		Config:          config,
		History:         transactions,
		Debug:           false,
		ResponseCounter: 0,
		Tabs:            make([]*godet.Tab, 0),
		Responses:       make(chan BrowserTransaction),
	}
}
