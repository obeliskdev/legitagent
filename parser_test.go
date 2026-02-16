package legitagent

import (
	"errors"
	"strings"
	"testing"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

func TestFromUserAgentString(t *testing.T) {
	t.Run("Successful Chrome Parse", func(t *testing.T) {
		ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36"
		agent, err := FromUserAgentString(ua, RequestTypeNavigate)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if agent.UserAgent != ua {
			t.Errorf("Expected UserAgent to be identical, got %s", agent.UserAgent)
		}
		if agent.ClientHelloID != utls.HelloChrome_133 {
			t.Error("Expected ClientHelloID for Chrome 138 to be HelloChrome_133")
		}
		if !strings.Contains(agent.Headers.Get("sec-ch-ua"), `"Google Chrome";v="138"`) {
			t.Errorf("sec-ch-ua header is incorrect: %s", agent.Headers.Get("sec-ch-ua"))
		}
	})

	t.Run("Closest Version Match", func(t *testing.T) {
		ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
		agent, err := FromUserAgentString(ua, RequestTypeNavigate)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if agent.ClientHelloID != utls.HelloChrome_120 {
			t.Error("Expected ClientHelloID for Chrome 125 to be HelloChrome_120")
		}
	})

	t.Run("Successful Firefox Parse", func(t *testing.T) {
		ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0"
		agent, err := FromUserAgentString(ua, RequestTypeNavigate)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if agent.Headers.Get("sec-ch-ua") != "" {
			t.Error("Firefox should not have sec-ch-ua headers")
		}
		if agent.ClientHelloID != utls.HelloFirefox_120 {
			t.Errorf("Incorrect TLS profile for Firefox, expected HelloFirefox_120")
		}
		if got := agent.H2Settings[http2.SettingInitialWindowSize]; got != GetGeckoH2Settings()[http2.SettingInitialWindowSize] {
			t.Errorf("Incorrect H2 profile for Firefox, expected gecko settings")
		}
	})

	t.Run("Successful Safari Mobile Parse", func(t *testing.T) {
		ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1"

		agent, err := FromUserAgentString(ua, RequestTypeNavigate)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if agent.UserAgent != ua {
			t.Errorf("Expected UserAgent to be identical, got %s", agent.UserAgent)
		}

		if agent.ClientHelloID != utls.HelloSafari_16_0 {
			t.Errorf("Incorrect TLS profile for Safari, expected HelloSafari_16_0")
		}
		if got := agent.H2Settings[http2.SettingMaxHeaderListSize]; got != GetWebKitH2Settings()[http2.SettingMaxHeaderListSize] {
			t.Errorf("Incorrect H2 profile for Safari, expected webkit settings")
		}
	})

	t.Run("Unsupported Browser", func(t *testing.T) {
		ua := "curl/7.64.1"
		_, err := FromUserAgentString(ua, RequestTypeNavigate)
		if !errors.Is(err, ErrUnsupportedBrowser) {
			t.Errorf("Expected ErrUnsupportedBrowser, got %v", err)
		}
	})

	t.Run("Unsupported Version", func(t *testing.T) {
		ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.0.0 Safari/537.36"
		_, err := FromUserAgentString(ua, RequestTypeNavigate)
		if !errors.Is(err, ErrUnsupportedVersion) {
			t.Errorf("Expected ErrUnsupportedVersion, got %v", err)
		}
	})

	t.Run("Android And ChromeOS Are Not Misclassified As Linux", func(t *testing.T) {
		androidUA := "Mozilla/5.0 (Linux; Android 14; Pixel 8 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Mobile Safari/537.36"
		androidParsed, err := parseUserAgentString(androidUA)
		if err != nil {
			t.Fatalf("Expected Android UA to parse, got %v", err)
		}
		if androidParsed.OS != OSAndroid {
			t.Fatalf("Expected Android OS, got %s", androidParsed.OS)
		}

		chromeOSUA := "Mozilla/5.0 (X11; CrOS x86_64 14541.0.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"
		chromeOSParsed, err := parseUserAgentString(chromeOSUA)
		if err != nil {
			t.Fatalf("Expected ChromeOS UA to parse, got %v", err)
		}
		if chromeOSParsed.OS != OSChromeOS {
			t.Fatalf("Expected ChromeOS OS, got %s", chromeOSParsed.OS)
		}
	})
}
