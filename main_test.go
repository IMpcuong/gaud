package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestDetectProxy(t *testing.T) {
	envs := syscall.Environ()
	for _, v := range envs {
		if !strings.Contains(strings.ToLower(v), "proxy") {
			fmt.Println("Not detected proxy configuration yet!")
		} else {
			fmt.Printf("Detected proxy configuration: %s\n", v)
		}
	}
}

func TestRegexProxy(t *testing.T) {
	regexProxy := regexp.MustCompile(`.+(proxy{1,2})=.*`)
	envVars := syscall.Environ()
	for _, v := range envVars {
		proxyEnv := string(regexProxy.Find([]byte(v)))
		if strings.Compare(proxyEnv, "") != 0 {
			fmt.Printf("Detected: %s\n", proxyEnv)
			host := strings.Split(proxyEnv, "=")[1]
			fmt.Printf("Host: %s\n", host)
			if strings.Contains(host, "//") {
				ipAddr := strings.Split(host, "://")[1]
				fmt.Printf("IP: %s\n", ipAddr)
				SOCK5_PROXY = strings.Join([]string{SOCK5_PROTO, ipAddr}, "://")
				fmt.Println("----->\t", SOCK5_PROXY)
				fmt.Println("----------")
			}
		}
	}
}

func TestExecPwshCmd(t *testing.T) {
	pwshOpts := []string{"-NoLogo", "-NoProfile", "-NonInteractive"}
	netSettingPath := "'" + filepath.Join(VERSION_SETTING, NET_SETTING) + "'"
	pwshArgs := fmt.Sprintf("Get-ItemProperty -Path %s", netSettingPath)
	fmt.Println(pwshArgs)

	// execution := exec.Command("powershell", pwshOpts[0], pwshOpts[1], pwshOpts[2], "Get-ItemProperty", "-Path", netSettingPath)
	execution := exec.Command("powershell", pwshOpts[0], pwshOpts[1], pwshOpts[2], pwshArgs)
	RunOSCmd(execution)
}

func TestNetSettings(t *testing.T) {
	// NOTE: Print 3 times because of the matched string's occurences in 3 output lines.
	settings := DetectNetSettings()
	fmt.Println(settings)
	settingsLowerCase := strings.ToLower(settings)
	if strings.Contains(settingsLowerCase, "proxy") {
		fmt.Println(settingsLowerCase)
	}
}

func TestIPLookUp(t *testing.T) {
	// lookUpLocalProxyDomain := IPLookUp("google.com")
	lookUpLocalProxyDomain := IPLookUp("genk.vn")
	fmt.Println("Local Proxy: \n", lookUpLocalProxyDomain)
}

func TestDetectSettingProxy(t *testing.T) {
	settings := DetectNetSettings()
	proxyCfg := DetectSettingProxy(settings)
	fmt.Println(proxyCfg)
}

func TestIsRangeSupported(t *testing.T) {
	settings := DetectNetSettings()
	proxyCfg := DetectSettingProxy(settings)
	client := NewHttpClient(proxyCfg, "", *new(time.Duration))

	isSupported := IsRangeSupported(client, BLOCKED_URL)
	fmt.Printf("Ranges supported URL: %s - %v\n", BLOCKED_URL, isSupported)
}

func TestCalculateRoutines(t *testing.T) {
	fmt.Println(CalculateChunk())
}

func TestSliceContains(t *testing.T) {
	slice := []string{"ProxyEnable", "MigrateProxy", "ProxyOverride", "AutoConfigURL", "ProxyServer"}
	if !SliceContains(slice, "ProxyServer") {
		log.Fatal("Error: Cannot detect required element!")
	}
	fmt.Printf("%#v\n", slice)
}

func TestSplitCarriageReturn(t *testing.T) {
	rawStr := "a\r\nb\r\nc\r\nd\r\ne"
	splitted := SplitCarriageReturn(rawStr)
	for _, s := range splitted {
		fmt.Println(s)
	}
}

func TestStringContains(t *testing.T) {
	raw := `a\r\nb\r\nc\nd\ne`
	if !strings.Contains(raw, `\r\n`) {
		log.Fatal("Error!")
	}
}
