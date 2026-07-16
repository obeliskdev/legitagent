package legitagent

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

var (
	ErrUnsupportedBrowser = errors.New("unsupported browser could not be parsed")
	ErrUnsupportedOS      = errors.New("unsupported os could not be parsed")
	ErrUnsupportedVersion = errors.New("no suitable profile found for the browser version")

	uaRegexes = []struct {
		Browser Browser
		Regex   *regexp.Regexp
	}{
		{BrowserSafari, regexp.MustCompile(`Version/(\d+)\..*Safari/`)},
		{BrowserEdge, regexp.MustCompile(`Edg(?:i|A)?OS?/(\d+)\.\d+`)},
		{BrowserOpera, regexp.MustCompile(`OPR/(\d+)\.\d+`)},
		{BrowserOpera, regexp.MustCompile(`OPT/(\d+)\.\d+`)},
		{BrowserBrave, regexp.MustCompile(`Brave/(\d+)\.\d+`)},
		{BrowserChrome, regexp.MustCompile(`(?:CriOS|Chrome)/(\d+)\.\d+`)},
		{BrowserFirefox, regexp.MustCompile(`(?:FxiOS|Firefox)/(\d+)\.\d+`)},
	}
)

var parserStableChromeProfiles = map[int]utls.ClientHelloID{
	120: utls.HelloChrome_120,
	131: utls.HelloChrome_131,
	133: utls.HelloChrome_133,
}
var parserStableChromeVersions []int

func init() {
	if len(parserStableChromeVersions) == 0 {
		for v := range parserStableChromeProfiles {
			parserStableChromeVersions = append(parserStableChromeVersions, v)
		}
		sort.Ints(parserStableChromeVersions)
	}
}

func findClosestChromeProfileForParser(targetVersion int) utls.ClientHelloID {
	bestVersion := -1
	for _, v := range parserStableChromeVersions {
		if v <= targetVersion && v > bestVersion {
			bestVersion = v
		}
	}
	if bestVersion != -1 {
		return parserStableChromeProfiles[bestVersion]
	}
	return parserStableChromeProfiles[parserStableChromeVersions[0]]
}

type parsedUA struct {
	Browser Browser
	Version int
	OS      OperatingSystem
}

func parseUserAgentString(ua string) (*parsedUA, error) {
	var p parsedUA

	for _, re := range uaRegexes {
		if match := re.Regex.FindStringSubmatch(ua); len(match) > 1 {
			p.Browser = re.Browser
			v, err := strconv.Atoi(match[1])
			if err != nil {
				return nil, fmt.Errorf("could not parse version from ua string: %w", err)
			}
			p.Version = v
			break
		}
	}
	if p.Browser == "" {
		return nil, ErrUnsupportedBrowser
	}

	p.OS = detectOSFromUA(ua, p.Browser)
	if p.OS == "" {
		return nil, ErrUnsupportedOS
	}

	return &p, nil
}

func detectOSFromUA(ua string, browser Browser) OperatingSystem {
	for i := 0; i < len(ua); {
		c := ua[i]
		switch c {
		case 'A':
			if strings.HasPrefix(ua[i:], "Android") {
				return OSAndroid
			}
			i++
		case 'C':
			if strings.HasPrefix(ua[i:], "CrOS") {
				return OSChromeOS
			}
			i++
		case 'i':
			if strings.HasPrefix(ua[i:], "iPhone") || strings.HasPrefix(ua[i:], "iPad") {
				return OSiOS
			}
			i++
		case 'W':
			if strings.HasPrefix(ua[i:], "Windows NT 10.0") {
				return OSWindows11
			}
			i++
		case 'M':
			if strings.HasPrefix(ua[i:], "Macintosh") {
				if browser == BrowserSafari && strings.Contains(ua, "Mobile/") {
					return OSiOS
				}
				return osMacIntel
			}
			i++
		case 'L':
			if strings.HasPrefix(ua[i:], "Linux; Android") {
				return OSAndroid
			}
			if strings.HasPrefix(ua[i:], "Linux") {
				return OSLinux
			}
			i++
		default:
			i++
		}
	}
	return ""
}

func FromUserAgentString(userAgentString string, requestType RequestType) (*Agent, error) {
	ua, err := parseUserAgentString(userAgentString)
	if err != nil {
		return nil, err
	}

	profile, ok := browserProfiles[ua.Browser]
	if !ok {
		return nil, ErrUnsupportedBrowser
	}

	osProf, ok := osProfiles[ua.OS]
	if !ok {
		return nil, ErrUnsupportedOS
	}

	platform := PlatformDesktop

	if ua.OS == OSAndroid || ua.OS == OSiOS {
		platform = PlatformMobile
	}

	platformProf, ok := platformProfiles[platform]
	if !ok {
		return nil, fmt.Errorf("internal error: no platform profile for %s", platform)
	}

	versionProf, _, err := findClosestVersionProfile(profile, ua.Version)
	if err != nil {
		return nil, err
	}

	fullVersion := ""
	if profile.ChromiumBased {
		fullVersion = strconv.Itoa(ua.Version) + ".0." + strconv.Itoa(versionProf.BuildNumber) + ".0"
	}

	agent := parserAgentPool.Get().(*Agent)
	header := agent.Headers
	if header == nil {
		header = make(http.Header, 16)
		agent.Headers = header
	}
	for k := range header {
		delete(header, k)
	}

	_, headerOrder := buildStaticHeaders(profile, osProf, platformProf, ua.Version, fullVersion, versionProf, requestType, header, agent.HeaderOrder)

	var helloID utls.ClientHelloID
	if profile.ChromiumBased {
		helloID = findClosestChromeProfileForParser(ua.Version)
	} else {
		helloID = versionProf.TLS.HelloID
		if helloID == (utls.ClientHelloID{}) {
			helloID = findClosestChromeProfileForParser(ua.Version)
		}
	}

	var h2Settings map[http2.SettingID]uint32
	var h2Pool *sync.Pool
	if profile.H2SettingsWithPool != nil {
		h2Settings, h2Pool = profile.H2SettingsWithPool()
	}

	agent.UserAgent = userAgentString
	agent.HeaderOrder = headerOrder
	agent.ClientHelloSpec = nil
	agent.ClientHelloID = helloID
	agent.H2Settings = h2Settings
	agent.h2SettingsPool = h2Pool

	return agent, nil
}

var parserAgentPool = sync.Pool{
	New: func() any {
		return new(Agent)
	},
}

func ReleaseParserAgent(a *Agent) {
	if a == nil {
		return
	}
	resetAgentFields(a)
	for k := range a.Headers {
		delete(a.Headers, k)
	}
	parserAgentPool.Put(a)
}

func findClosestVersionProfile(profile browserProfile, targetVersion int) (versionProfile, int, error) {
	closestVersion := -1
	for _, v := range profile.VersionKeys {
		if v <= targetVersion && v > closestVersion {
			closestVersion = v
		}
	}

	if closestVersion == -1 {
		return versionProfile{}, 0, ErrUnsupportedVersion
	}

	return profile.Versions[closestVersion], closestVersion, nil
}

func buildStaticHeaders(browser browserProfile, os osProfile, platform platformProfile, version int, fullVersion string, versionProf versionProfile, requestType RequestType, header http.Header, existingOrder []string) (http.Header, []string) {

	var acceptTemplate []AcceptHeaderPart
	switch requestType {
	case RequestTypeXHR:
		if len(versionProf.AcceptHeaderPatternsXHR) > 0 {
			acceptTemplate = versionProf.AcceptHeaderPatternsXHR[0]
		}
	default:
		if len(versionProf.AcceptHeaderPatterns) > 0 {
			acceptTemplate = versionProf.AcceptHeaderPatterns[0]
		}
	}

	headerSet(header, hAccept, buildAcceptHeaderStatic(acceptTemplate))
	headerSet(header, hAcceptEncoding, "gzip, deflate, br")
	headerSet(header, hAcceptLanguage, "en-US,en;q=0.9")

	if browser.ChromiumBased {
		headerSet(header, hSecChUa, buildSecChUa(browser.Brand, strconv.Itoa(version), false, false))
		headerSet(header, hSecChUaMobile, platform.MobileHint)
		headerSet(header, hSecChUaPlatform, os.PlatformQuote)
		headerSet(header, hSecChUaFullVersionList, buildSecChUa(browser.Brand, fullVersion, true, false))
		if os.PlatformVersionQ != "" {
			headerSet(header, hSecChUaPlatformVersion, os.PlatformVersionQ)
		}
		if os.ArchQuote != "" {
			headerSet(header, hSecChUaArch, os.ArchQuote)
		}
		if os.BitnessQuote != "" {
			headerSet(header, hSecChUaBitness, os.BitnessQuote)
		}
	}

	switch requestType {
	case RequestTypeNavigate:
		headerSet(header, hSecFetchDest, "document")
		headerSet(header, hSecFetchMode, "navigate")
		headerSet(header, hSecFetchSite, "none")
		headerSet(header, hSecFetchUser, "?1")
		headerSet(header, hUpgradeInsecureRequests, "1")
	case RequestTypeXHR:
		headerSet(header, hSecFetchDest, "empty")
		headerSet(header, hSecFetchMode, "cors")
		headerSet(header, hSecFetchSite, "same-origin")
	}

	keysPtr := keysPool.Get().(*[]string)
	keys := (*keysPtr)[:0]
	for k := range header {
		keys = append(keys, k)
	}

	PriorityHeaderSorter(keys)
	orderedKeys := rebuildHeaderOrder(existingOrder, keys)

	*keysPtr = keys
	keysPool.Put(keysPtr)

	return header, orderedKeys
}

func buildAcceptHeaderStatic(parts []AcceptHeaderPart) string {
	sb := builderPool.Get().(*strings.Builder)
	defer func() {
		sb.Reset()
		builderPool.Put(sb)
	}()

	for i, part := range parts {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(part.Value)
		for _, extra := range part.Extras {
			sb.WriteString(";")
			sb.WriteString(extra)
		}
		if part.Q > 0 {
			sb.WriteString(";q=")
			sb.WriteString(strconv.FormatFloat(part.Q, 'f', 1, 64))
		}
	}
	return sb.String()
}
