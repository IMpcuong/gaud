package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

/*
	Expected: `go main.go -d (<URL>)`.
	Or using the scripts as below:
		+ Local environment: `./scripts/run.sh local (<URL>)`.
		+ Docker container: `./scripts/run.sh docker (<URL>)`.

	NOTE: Value inside brackets is optional, if the URL was not specified, using `TEST_URL` instead.
*/

const (
	TEST_URL    string = "https://yt3.ggpht.com/a/AATXAJy_8AbZ24NBUacQ_EGRdK3a1y11VCV4mF_ID-jOAw=s900-c-k-c0xffffffff-no-rj-mo"
	NORMAL_URL  string = "https://github.com/microsoft/terminal/releases/download/v1.15.2282.0/Microsoft.WindowsTerminalPreview_Win10_1.15.2282.0_8wekyb3d8bbwe.msixbundle"
	BLOCKED_URL string = "https://github.com/kubernetes/minikube/releases/download/v1.26.1/minikube-windows-amd64.exe"
)

func main() {
	var url string

	flag.StringVar(&url, "d", "", "Specify URL to download. Default is TEST_URL.")
	flag.Usage = func() {
		fmt.Printf("Usage guidance for our GAD Download Application: \n")
		fmt.Printf("+ Windows: ./gad.exe -d <URI>\n")
		fmt.Printf("+ Linux: ./gad -d <URI>\n\n")

		fmt.Printf("Options:\n")
		flag.PrintDefaults() // Prints default usage.
	}
	flag.Parse()

	if strings.Compare(TEST_URL, url) != 0 {
		DownloadFrom(url)
		return
	}
	DownloadFrom(TEST_URL)
}

// TODO: Maybe we will use this abstraction object, or maybe not.
// Major/vital/prime object to execute the downloading process using proxy-tunnel.
type Gad struct {
	Thread    int
	URI       string
	Chunks    map[int]*os.File
	StartTime time.Time
	FileName  string
	OutFile   *os.File
	Err       error
	*sync.Mutex
}

func DownloadFrom(rawPath string) {
	// NOTE: Read path from stdin in terminal.
	cmd := exec.Command(rawPath)

	// Build fileName from full-path URL:
	segments := SplitUrl(rawPath)
	fileName := segments[len(segments)-1]

	// Create a blank file:
	fileStore := CreateFile(fileName)

	// Request to download a specific file from existed URL:
	chunks := strconv.Itoa(CalculateChunk())
	resp := ReqUrl(rawPath, chunks)
	defer resp.Body.Close()

	// Put content to new file:
	content := resp.Body
	size, err := io.Copy(fileStore, content)
	HandleError(err)
	fmt.Printf("Downloaded a file %s with size %d\n", fileName, size)

	defer fileStore.Close()

	RunOSCmd(cmd)
}

var SOCK5_PROXY string

func ReqUrl(url string, ranges string) *http.Response {
	proxy := DetectEnvProxy()
	if strings.Compare(proxy, "") == 0 {
		settings := DetectNetSettings()
		proxy = DetectSettingProxy(settings)
	}
	fmt.Printf("Proxy configuration: `%s`\n", proxy)

	// Generate new http.client:
	timeout := new(time.Duration)
	client := NewHttpClient(proxy, ranges, *timeout)

	var resp *http.Response
	var err error
	if IsRangeSupported(client, url) {
		// Download file from given URL:
		resp, err = client.Get(url)
		HandleError(err)

		return resp
	}

	return resp
}

func NewHttpClient(proxy, chunk string, sec time.Duration) http.Client {
	var client http.Client

	if strings.Compare(chunk, "") == 0 {
		client = http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				req.URL.Opaque = req.URL.Path
				return nil
			},
			Transport: &http.Transport{
				Proxy: http.ProxyURL(UrlConverter(proxy)),
			},
			Timeout: sec,
		}
	}

	rangeReq := fmt.Sprintf("%s=%s", "bytes", chunk)
	client = http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.Header.Add("Range", rangeReq)
			req.URL.Opaque = req.URL.Path
			return nil
		},
		Transport: &http.Transport{
			Proxy: http.ProxyURL(UrlConverter(proxy)),
		},
		Timeout: sec,
	}

	return client
}

func IsRangeSupported(client http.Client, url string) bool {
	resp, err := client.Head(url)
	HandleError(err)

	header := resp.Header
	strContentLen := header.Get("Content-Length")
	contentLen, err := strconv.Atoi(strContentLen)
	HandleError(err)

	if contentLen == 0 {
		log.Fatal("Error: Your wished file to download was totally empty!")
		return false
	}

	if strings.Compare(header.Get("Accept-Ranges"), "bytes") == 0 {
		return true
	}
	return false
}

func UrlConverter(raw string) *url.URL {
	if strings.Compare(raw, "") == 0 {
		log.Fatal("Error: Cannot parse null/nil string to URL!")
	}
	converted, err := url.Parse(raw)
	HandleError(err)

	return converted
}

func SplitUrl(url string) []string {
	if strings.Compare(url, "") == 0 {
		log.Fatal("Error: Cannot split null/nil URL to an array!")
	}

	if !strings.Contains(url, "/") {
		log.Fatal("Error: URL formatter mismatch!")
	}

	fullUrl := UrlConverter(url)

	// Split path to retrieve `fileName` only:
	return strings.Split(fullUrl.Path, "/")
}

func CreateFile(fileName string) *os.File {
	newFile, err := os.Create(fileName)
	HandleError(err)

	return newFile
}

func RunOSCmd(cmd *exec.Cmd) {
	// Mapping OS's standard streams with commandline interface:
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}

const (
	SOCK5_PROTO     = "socks5"
	HTTPS_PROTO     = "https"
	HTTP_PROTO      = "http"
	NET_SETTING     = "Internet Settings"
	VERSION_SETTING = "HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion"
)

func DetectEnvProxy() string {
	regexProxy := regexp.MustCompile(`.+(proxy{1})=.*`)

	envVars := syscall.Environ()
	for _, v := range envVars {
		proxyEnv := string(regexProxy.Find([]byte(v)))
		if strings.Compare(proxyEnv, "") == 0 {
			continue
		}
		host := strings.Split(proxyEnv, "=")[1]
		if strings.Contains(host, "//") {
			// NOTE: `[1]` position in the array represents for the right-hand side value against the colon.
			ipAddr := strings.Split(host, "://")[1]
			SOCK5_PROXY = strings.Join([]string{SOCK5_PROTO, ipAddr}, "://")
			return SOCK5_PROXY
		}
	}
	return SOCK5_PROXY
}

// Store StandardOutput payload to a variable.
type StdoutStore struct {
	Data []byte
}

func (ss *StdoutStore) Write(payload []byte) (n int, err error) {
	ss.Data = append(ss.Data, payload...)
	return os.Stdout.Write(payload)
}

func (ss *StdoutStore) ExecPwshCmd(arg string) string {
	if strings.Compare(arg, "") == 0 {
		log.Fatal("Error: Arguments to execute command is required!")
	}

	var stdOut string
	if runtime.GOOS != "windows" {
		return stdOut
	}

	pwshOpts := []string{"-NoLogo", "-NoProfile", "-NonInteractive"}
	execution := exec.Command("powershell", pwshOpts[0], pwshOpts[1], pwshOpts[2], arg)
	execution.Stdin = os.Stdin
	execution.Stdout = ss
	execution.Stderr = os.Stderr
	_ = execution.Run()

	stdOut = string(ss.Data)
	return stdOut
}

func DetectNetSettings() string {
	var settings string
	var ss StdoutStore

	netSettingPath := "'" + filepath.Join(VERSION_SETTING, NET_SETTING) + "'"
	pwshArgs := fmt.Sprintf("Get-ItemProperty -Path %s", netSettingPath)

	settings = string(ss.ExecPwshCmd(pwshArgs))
	return settings
}

const (
	// From: https://stackoverflow.com/questions/5284147/validating-ipv4-addresses-with-regexp
	IPv4_PATTERN string = `^.*((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}.*$`
	// Read more: https://en.wikipedia.org/wiki/Carriage_return
	CARRIAGE_RETURN      string = `\r\n`
	SPECIAL_CHARS_DOMAIN string = `[^\w\.\-]`
)

var (
	regexIPv4         = regexp.MustCompile(IPv4_PATTERN)
	regexSpecialChars = regexp.MustCompile(SPECIAL_CHARS_DOMAIN)
)

func DetectSettingProxy(settings string) string {
	var netSetting []string
	var proxy string

	settingsLowerCase := strings.ToLower(settings)
	netSetting = SplitCarriageReturn(settingsLowerCase)
	for _, setting := range netSetting {
		if !strings.Contains(setting, "proxy") {
			continue
		}

		// NOTE: Pattern equivalent with a compound of a colon and a space character.
		proxyCfgs := strings.Split(setting, ": ")[:]
		// fmt.Printf("%#v - %d\n", proxyCfgs, len(proxyCfgs))

		// NOTE: The left-hand side equals to := `[]string{"ProxyEnable", "MigrateProxy", "ProxyOverride", "AutoConfigURL", "ProxyServer"}`.
		switch strings.TrimSpace(proxyCfgs[0]) {
		// NOTE: This configuration below only have higher priority in our organization only.
		// NOTE: `[1]` position in the array represents for the right-hand side value against the colon.
		case "autoconfigurl":
			// NOTE: pattern := `http://<proxy-domain>:<proxy-port>/proxy.pac`.
			proxyCfg := strings.Split(proxyCfgs[1], ":")
			proxyDomain := proxyCfg[1]
			proxyPort := strings.Split(proxyCfg[2], "/")[0]

			badChars := regexSpecialChars.FindAllString(proxyDomain, -1)
			if len(badChars) == 0 {
				proxy = fmt.Sprintf("%s:%s", IPLookUp(proxyDomain), proxyPort)
				CraftSock5Proxy(proxy)
			}

			for _, char := range badChars {
				proxyDomain = strings.Trim(proxyDomain, char)
			}
			proxy = fmt.Sprintf("%s:%s", IPLookUp(proxyDomain), proxyPort)
			CraftSock5Proxy(proxy)
		case "proxyserver":
			proxy = string(regexIPv4.Find([]byte(proxyCfgs[1])))
			CraftSock5Proxy(proxy)
		}
	}

	return SOCK5_PROXY
}

func SplitCarriageReturn(mixed string) []string {
	if strings.Compare(mixed, "") == 0 {
		log.Fatal("Error: Null/nil string was not allowed!")
	}

	var splittedStr []string
	rawStr := fmt.Sprintf("%#v", mixed)
	if !strings.Contains(rawStr, CARRIAGE_RETURN) {
		log.Fatal("Error: The given string doesn't contain carriage return character!")
	}

	splittedStr = append(splittedStr, strings.Split(rawStr, CARRIAGE_RETURN)...)
	return splittedStr
}

func IPLookUp(domain string) string {
	if strings.Compare(domain, "") == 0 {
		log.Fatal("Error: Null/nil domain's name was not allowed!")
	}

	var strIPv4 string
	var ss StdoutStore
	pwshArgs := fmt.Sprintf("nslookup %s", domain)
	serverInfo := string(ss.ExecPwshCmd(pwshArgs))
	serverInfoLowerCase := strings.ToLower(serverInfo)

	var listIPs []string
	addresses := SplitCarriageReturn(serverInfoLowerCase)
	// fmt.Printf("%#v\n", addresses)
	for _, addr := range addresses {
		// FIXME: The concept of multiple IP addresses were binding into one domain make thing become even more spooky.
		// Solution: For now, we will only take our focus on the second IPv4 was appeared in the output result.
		mixedIPv4 := regexIPv4.FindString(addr)
		if strings.Compare(mixedIPv4, "") == 0 {
			continue
		}

		// NOTE: The list of all correctness formatter for IPv4 string.
		listIPs = append(listIPs, mixedIPv4)
	}

	if len(listIPs) == 1 {
		log.Fatal("Error: Your local machine doesn't configure proxy. Please download manually!")
	}

	// NOTE: The second (position 1 in the slice) IPv4 address is the one that we was looking for (the underlying IP of the proxy-domain).
	// NOTE: `[1]` position in the array represents for the right-hand side value against the colon.
	strIPv4 = strings.Split(listIPs[1], ":")[1]
	return strings.TrimSpace(strIPv4)
}

func CraftSock5Proxy(proxy string) string {
	if strings.Compare(proxy, "") == 0 {
		log.Fatal("Error: Proxy-configurations cannot be null/nil!")
	}

	if strings.Contains(proxy, "//") {
		SOCK5_PROXY = fmt.Sprintf("%s:%s", SOCK5_PROTO, proxy)
		return SOCK5_PROXY
	}

	SOCK5_PROXY = fmt.Sprintf("%s://%s", SOCK5_PROTO, proxy)
	return SOCK5_PROXY
}

func SliceContains[T comparable](slice []T, checked T) bool {
	for _, val := range slice {
		if val == checked {
			return true
		}
	}
	return false
}

// FIXME: Using magic numbers `10000` in here, need to findout a better solution to calculate this number.
func CalculateChunk() int {
	return runtime.NumCPU() * 10000
}

func HandleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
