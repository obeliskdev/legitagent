package legitagent

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/obeliskdev/fastrand"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

type AcceptHeaderPart struct {
	Value  string
	Q      float64
	Extras []string
}

type UAComponentGenerator func(browserProfile browserProfile, osProfile osProfile, versionProfile versionProfile, fullVersion string) string
type BrowserFamily string

const (
	Chromium BrowserFamily = "Chromium"
	Gecko    BrowserFamily = "Gecko"
	WebKit   BrowserFamily = "WebKit"
)

type tlsProfile struct {
	HelloID    utls.ClientHelloID
	ClientSpec func() *utls.ClientHelloSpec
}

type versionProfile struct {
	BuildNumber             int
	AcceptHeaderPatterns    [][]AcceptHeaderPart
	AcceptHeaderPatternsXHR [][]AcceptHeaderPart
	TLS                     tlsProfile
	GeckoRevision           string
	WebKitVersion           string
	MobileVersion           string
	SafariVersion           string
	SupportsH2              bool
}

type browserProfile struct {
	Brand              string
	Family             BrowserFamily
	UASuffix           string
	Versions           map[int]versionProfile
	VersionKeys        []int
	H2VersionKeys      []int
	ChromiumBased      bool
	H2Settings         func() map[http2.SettingID]uint32
	H2SettingsWithPool func() (map[http2.SettingID]uint32, *sync.Pool)
}

type osProfile struct {
	Name             string
	PlatformToken    string
	Version          string
	Arch             string
	BitnessHint      string
	IsMobile         bool
	IsMacOS          bool
	PlatformQuote    string
	PlatformVersionQ string
	ArchQuote        string
	BitnessQuote     string
}

type platformProfile struct {
	MobileHint          string
	ComponentGenerators map[BrowserFamily][]UAComponentGenerator
}

var (
	acceptHeaderPatternsChrome = [][]AcceptHeaderPart{
		{
			{Value: "text/html"},
			{Value: "application/xhtml+xml"},
			{Value: "application/xml", Q: 0.9},
			{Value: "image/avif"},
			{Value: "image/webp"},
			{Value: "image/apng"},
			{Value: "*/*", Q: 0.8},
			{Value: "application/signed-exchange", Q: 0.7, Extras: []string{"v=b3"}},
		},
		{
			{Value: "text/html"},
			{Value: "application/xhtml+xml"},
			{Value: "application/xml", Q: 0.9},
			{Value: "image/avif"},
			{Value: "image/webp"},
			{Value: "image/apng"},
			{Value: "*/*", Q: 0.8},
		},
	}
	acceptHeaderPatternsFirefox = [][]AcceptHeaderPart{
		{
			{Value: "text/html"},
			{Value: "application/xhtml+xml"},
			{Value: "application/xml", Q: 0.9},
			{Value: "image/avif"},
			{Value: "image/webp"},
			{Value: "*/*", Q: 0.8},
		},
	}
	acceptHeaderPatternsSafari = [][]AcceptHeaderPart{
		{
			{Value: "text/html"},
			{Value: "application/xhtml+xml"},
			{Value: "application/xml", Q: 0.9},
			{Value: "*/*", Q: 0.8},
		},
	}
	acceptHeaderPatternsXHR = [][]AcceptHeaderPart{
		{{Value: "*/*"}},
	}

	tlsProfileChrome120  = tlsProfile{HelloID: utls.HelloChrome_120}
	tlsProfileChrome131  = tlsProfile{HelloID: utls.HelloChrome_131}
	tlsProfileChrome133  = tlsProfile{HelloID: utls.HelloChrome_133}
	tlsProfileFirefox120 = tlsProfile{HelloID: utls.HelloFirefox_120}
	tlsProfileSafari16   = tlsProfile{HelloID: utls.HelloSafari_16_0}

	chromeVersions = map[int]versionProfile{
		114: {BuildNumber: 5735, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome120, SupportsH2: true},
		116: {BuildNumber: 5845, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome120, SupportsH2: true},
		118: {BuildNumber: 5993, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome120, SupportsH2: true},
		120: {BuildNumber: 6099, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome120, SupportsH2: true},
		124: {BuildNumber: 6367, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome131, SupportsH2: true},
		128: {BuildNumber: 6613, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome131, SupportsH2: true},
		130: {BuildNumber: 6723, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome131, SupportsH2: true},
		133: {BuildNumber: 6943, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		140: {BuildNumber: 7339, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		141: {BuildNumber: 7390, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		145: {BuildNumber: 7632, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		146: {BuildNumber: 7680, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		147: {BuildNumber: 7727, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		148: {BuildNumber: 7778, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		149: {BuildNumber: 7827, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		150: {BuildNumber: 7871, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		151: {BuildNumber: 7922, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
	}
	edgeVersions = map[int]versionProfile{
		114: {BuildNumber: 1823, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome120, SupportsH2: true},
		116: {BuildNumber: 1938, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome120, SupportsH2: true},
		118: {BuildNumber: 2088, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome120, SupportsH2: true},
		120: {BuildNumber: 2210, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome120, SupportsH2: true},
		124: {BuildNumber: 2478, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome131, SupportsH2: true},
		128: {BuildNumber: 2739, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome131, SupportsH2: true},
		130: {BuildNumber: 2835, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome131, SupportsH2: true},
		133: {BuildNumber: 2988, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		140: {BuildNumber: 3265, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		141: {BuildNumber: 3537, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		145: {BuildNumber: 3800, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		146: {BuildNumber: 3928, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		147: {BuildNumber: 4048, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		148: {BuildNumber: 4187, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		149: {BuildNumber: 4263, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		150: {BuildNumber: 4283, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
		151: {BuildNumber: 4318, AcceptHeaderPatterns: acceptHeaderPatternsChrome, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileChrome133, SupportsH2: true},
	}
	braveVersions = chromeVersions

	browserProfiles = map[Browser]browserProfile{
		BrowserChrome: {Brand: "Google Chrome", Family: Chromium, UASuffix: "", ChromiumBased: true, Versions: chromeVersions, H2Settings: GetChromiumH2Settings, H2SettingsWithPool: GetChromiumH2SettingsWithPool},
		BrowserOpera:  {Brand: "Opera", Family: Chromium, UASuffix: "OPR/%s", ChromiumBased: true, Versions: chromeVersions, H2Settings: GetChromiumH2Settings, H2SettingsWithPool: GetChromiumH2SettingsWithPool},
		BrowserEdge:   {Brand: "Microsoft Edge", Family: Chromium, UASuffix: "Edg/%s", ChromiumBased: true, Versions: edgeVersions, H2Settings: GetChromiumH2Settings, H2SettingsWithPool: GetChromiumH2SettingsWithPool},
		BrowserBrave:  {Brand: "Brave", Family: Chromium, UASuffix: "", ChromiumBased: true, Versions: braveVersions, H2Settings: GetChromiumH2Settings, H2SettingsWithPool: GetChromiumH2SettingsWithPool},
		BrowserFirefox: {Brand: "Firefox", Family: Gecko, ChromiumBased: false, Versions: map[int]versionProfile{
			115: {GeckoRevision: "115.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			120: {GeckoRevision: "120.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			127: {GeckoRevision: "127.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			128: {GeckoRevision: "128.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			130: {GeckoRevision: "130.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			131: {GeckoRevision: "131.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			133: {GeckoRevision: "133.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			134: {GeckoRevision: "134.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			135: {GeckoRevision: "135.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			136: {GeckoRevision: "136.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			137: {GeckoRevision: "137.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			138: {GeckoRevision: "138.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			139: {GeckoRevision: "139.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			140: {GeckoRevision: "140.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
			141: {GeckoRevision: "141.0", AcceptHeaderPatterns: acceptHeaderPatternsFirefox, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileFirefox120, SupportsH2: true},
		}, H2Settings: GetGeckoH2Settings, H2SettingsWithPool: GetGeckoH2SettingsWithPool},
		BrowserSafari: {Brand: "Safari", Family: WebKit, ChromiumBased: false, Versions: map[int]versionProfile{
			16: {WebKitVersion: "605.1.15", MobileVersion: "20F66", SafariVersion: "16.5", AcceptHeaderPatterns: acceptHeaderPatternsSafari, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileSafari16, SupportsH2: true},
			17: {WebKitVersion: "605.1.15", MobileVersion: "15E148", SafariVersion: "17.5", AcceptHeaderPatterns: acceptHeaderPatternsSafari, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileSafari16, SupportsH2: true},
			18: {WebKitVersion: "605.1.15", MobileVersion: "22B92", SafariVersion: "18.1", AcceptHeaderPatterns: acceptHeaderPatternsSafari, AcceptHeaderPatternsXHR: acceptHeaderPatternsXHR, TLS: tlsProfileSafari16, SupportsH2: true},
		}, H2Settings: GetWebKitH2Settings, H2SettingsWithPool: GetWebKitH2SettingsWithPool},
	}

	osProfiles = map[OperatingSystem]osProfile{
		OSWindows:         {Name: "Windows", PlatformToken: "Windows NT 10.0; Win64; x64", Version: "10.0.0", Arch: "x86", BitnessHint: "64", IsMobile: false},
		OSWindows11:       {Name: "Windows", PlatformToken: "Windows NT 10.0; Win64; x64", Version: "15.0.0", Arch: "x86", BitnessHint: "64", IsMobile: false},
		osMacIntel:        {Name: "macOS", PlatformToken: "Macintosh; Intel Mac OS X 10_15_7", Version: "14.5.0", Arch: "x86", BitnessHint: "64", IsMobile: false},
		osMacAppleSilicon: {Name: "macOS", PlatformToken: "Macintosh; Intel Mac OS X 10_15_7", Version: "14.5.0", Arch: "arm", BitnessHint: "64", IsMobile: false},
		OSLinux:           {Name: "Linux", PlatformToken: "X11; Linux x86_64", Version: "", Arch: "x86", BitnessHint: "64", IsMobile: false},
		osUbuntu:          {Name: "Linux", PlatformToken: "X11; Ubuntu; Linux x86_64", Version: "", Arch: "x86", BitnessHint: "64", IsMobile: false},
		osFedora:          {Name: "Linux", PlatformToken: "X11; Fedora; Linux x86_64", Version: "", Arch: "x86", BitnessHint: "64", IsMobile: false},
		OSAndroid:         {Name: "Android", PlatformToken: "Linux; Android 15; {device_model}", Version: "15.0.0", Arch: "arm", BitnessHint: "64", IsMobile: true},
		OSiOS:             {Name: "iOS", PlatformToken: "iPhone; CPU iPhone OS 18_5 like Mac OS X", Version: "18.5.0", IsMobile: true},
		OSChromeOS:        {Name: "Chrome OS", PlatformToken: "X11; CrOS x86_64 15923.0.0", Version: "15923.0.0", Arch: "x86", BitnessHint: "64", IsMobile: false},
	}

	platformProfiles = map[Platform]platformProfile{
		PlatformDesktop: {MobileHint: "?0", ComponentGenerators: map[BrowserFamily][]UAComponentGenerator{
			Chromium: {MozillaGenerator, OSGenerator, WebKitGenerator, KHTMLGenerator, ChromeGenerator, SafariGenerator, BrowserSuffixGenerator},
			Gecko:    {MozillaGenerator, FirefoxOSGenerator, GeckoTrailGenerator, FirefoxVersionGenerator},
			WebKit:   {MozillaGenerator, OSGenerator, SafariWebKitGenerator, KHTMLGenerator, SafariVersionGenerator, SafariBrowserVersionGenerator},
		}},
		PlatformMobile: {MobileHint: "?1", ComponentGenerators: map[BrowserFamily][]UAComponentGenerator{
			Chromium: {MozillaGenerator, OSGenerator, WebKitGenerator, KHTMLGenerator, ChromeGenerator, MobileSafariGenerator, BrowserSuffixGenerator},
			Gecko:    {MozillaGenerator, FirefoxOSGenerator, GeckoTrailGenerator, FirefoxVersionGenerator},
			WebKit:   {MozillaGenerator, OSGenerator, SafariWebKitGenerator, KHTMLGenerator, SafariVersionGenerator, SafariMobileTokenGenerator, SafariBrowserVersionGenerator},
		}},
	}
)

type greaseBrandParts struct {
	Key     string
	Version string
}

var greaseBrandParsed = []greaseBrandParts{
	{Key: `"Not/A)Brand"`, Version: `"8"`},
	{Key: `"Not;A Brand"`, Version: `"99"`},
	{Key: `"Not(A:Brand"`, Version: `"24"`},
}

var greaseBrandDefault = greaseBrandParsed[0]

var androidDevices = []string{"Pixel 8", "Pixel 8 Pro", "Pixel 9", "Pixel 9 Pro", "SM-S928B", "SM-S938B", "SM-G991U", "SM-F936U", "2201116SG", "V2109", "SM-A525F", "Pixel 6a", "SM-A536U", "SM-S926U", "SM-S931U", "CPH2615", "CPH2581"}
var subresourceDests = []string{"style", "script", "image", "font", "empty"}

func init() {
	for k, bp := range browserProfiles {
		keys := make([]int, 0, len(bp.Versions))
		for v := range bp.Versions {
			keys = append(keys, v)
		}
		sort.Ints(keys)
		bp.VersionKeys = keys

		h2keys := make([]int, 0, len(keys))
		for _, v := range keys {
			if bp.Versions[v].SupportsH2 {
				h2keys = append(h2keys, v)
			}
		}
		bp.H2VersionKeys = h2keys

		browserProfiles[k] = bp
	}

	for k, op := range osProfiles {
		op.PlatformQuote = `"` + op.Name + `"`
		op.IsMacOS = op.Name == "macOS"
		if op.Version != "" {
			op.PlatformVersionQ = `"` + op.Version + `"`
		}
		if op.Arch != "" {
			op.ArchQuote = `"` + op.Arch + `"`
		}
		if op.BitnessHint != "" {
			op.BitnessQuote = `"` + op.BitnessHint + `"`
		}
		osProfiles[k] = op
	}
}

func MozillaGenerator(_ browserProfile, _ osProfile, _ versionProfile, _ string) string {
	return "Mozilla/5.0"
}
func KHTMLGenerator(_ browserProfile, _ osProfile, _ versionProfile, _ string) string {
	return "(KHTML, like Gecko)"
}
func ChromeGenerator(_ browserProfile, _ osProfile, _ versionProfile, fv string) string {
	return "Chrome/" + fv
}
func SafariGenerator(_ browserProfile, _ osProfile, _ versionProfile, _ string) string {
	return "Safari/537.36"
}
func MobileSafariGenerator(_ browserProfile, _ osProfile, _ versionProfile, _ string) string {
	return "Mobile Safari/537.36"
}
func GeckoTrailGenerator(_ browserProfile, _ osProfile, _ versionProfile, _ string) string {
	return "Gecko/20100101"
}

func resolveDeviceToken(op osProfile) string {
	token := op.PlatformToken
	if op.Name == "Android" {
		device := fastrand.Choice(androidDevices)
		token = strings.Replace(token, "{device_model}", device, 1)
	}
	return token
}

func OSGenerator(_ browserProfile, op osProfile, _ versionProfile, _ string) string {
	return "(" + resolveDeviceToken(op) + ")"
}

func FirefoxOSGenerator(_ browserProfile, op osProfile, vp versionProfile, _ string) string {
	return "(" + resolveDeviceToken(op) + "; rv:" + vp.GeckoRevision + ")"
}

func FirefoxVersionGenerator(_ browserProfile, _ osProfile, vp versionProfile, _ string) string {
	return "Firefox/" + vp.GeckoRevision
}

func WebKitGenerator(_ browserProfile, _ osProfile, _ versionProfile, _ string) string {
	return "AppleWebKit/537.36"
}

func SafariWebKitGenerator(_ browserProfile, _ osProfile, vp versionProfile, _ string) string {
	return "AppleWebKit/" + vp.WebKitVersion
}

func SafariVersionGenerator(_ browserProfile, _ osProfile, vp versionProfile, _ string) string {
	return "Version/" + vp.SafariVersion
}

func SafariMobileTokenGenerator(_ browserProfile, _ osProfile, vp versionProfile, _ string) string {
	return "Mobile/" + vp.MobileVersion
}

func SafariBrowserVersionGenerator(_ browserProfile, op osProfile, vp versionProfile, _ string) string {
	if op.IsMobile {
		return "Safari/604.1"
	}
	return "Safari/" + vp.WebKitVersion
}

func BrowserSuffixGenerator(bp browserProfile, _ osProfile, _ versionProfile, fv string) string {
	if bp.UASuffix == "" {
		return ""
	}
	majorVersion := fv
	if idx := strings.IndexByte(fv, '.'); idx >= 0 {
		majorVersion = fv[:idx]
	}
	return fmt.Sprintf(bp.UASuffix, majorVersion)
}
