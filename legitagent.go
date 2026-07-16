package legitagent

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/obeliskdev/fastrand"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

var (
	ErrNoBrowsers         = errors.New("legitagent: no browsers configured for generation")
	ErrNoBotProfiles      = errors.New("legitagent: no bot profiles found for the specified types")
	ErrNoVersions         = errors.New("legitagent: no available browser versions that meet the specified criteria")
	ErrNoLanguageProfiles = errors.New("legitagent: no language profiles configured")
	ErrNoPlatformOSCombo  = errors.New("legitagent: no compatible platform/OS combination found")
)

type Agent struct {
	UserAgent       string
	Headers         http.Header
	HeaderOrder     []string
	ClientHelloSpec *utls.ClientHelloSpec
	ClientHelloID   utls.ClientHelloID
	H2Settings      map[http2.SettingID]uint32
	h2SettingsPool  *sync.Pool
}

type Generator struct {
	browsers               []Browser
	platforms              []Platform
	os                     []OperatingSystem
	minVersion             int
	maxVersion             int
	languageProfiles       [][]AcceptHeaderPart
	requestType            RequestType
	headerSorter           HeaderSorter
	fullFingerprint        bool
	h2Only                 bool
	fingerprintProfile     FingerprintProfile
	h2RandomizationProfile H2RandomizationProfile
	useBotAgents           bool
	botAgentTypes          []string
	acceptEncodingEnabled  bool
	acceptEnabled          bool
	agentPool              sync.Pool
	zeroHeader             bool
}

var allRealOS = []OperatingSystem{
	OSWindows,
	OSWindows11,
	OSLinux,
	OSMac,
	OSAndroid,
	OSiOS,
	OSChromeOS,
	osUbuntu,
	osFedora,
}
var allRealBrowsers = []Browser{
	BrowserChrome,
	BrowserOpera,
	BrowserEdge,
	BrowserBrave,
	BrowserFirefox,
	BrowserSafari,
}
var allRealPlatforms = []Platform{
	PlatformDesktop,
	PlatformMobile,
}
var macArchitectures = []OperatingSystem{
	osMacIntel,
	osMacAppleSilicon,
}

const (
	defaultMinChromiumVersion = 114
	defaultMaxChromiumVersion = 151
)

var builderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

var keysPool = sync.Pool{
	New: func() interface{} {
		s := make([]string, 0, 24)
		return &s
	},
}

var comboPool = sync.Pool{
	New: func() interface{} {
		s := make([]platformOSCombo, 0, 20)
		return &s
	},
}

var pseudoHeaders = []string{":method", ":authority", ":scheme", ":path"}

var defaultLanguages = [][]AcceptHeaderPart{
	{{Value: "en-US"}, {Value: "en", Q: 0.9}},
	{{Value: "de-DE"}, {Value: "de", Q: 0.9}},
	{{Value: "fa-IR"}, {Value: "fa", Q: 0.9}},
	{{Value: "fr-FR"}, {Value: "fr", Q: 0.9}},
	{{Value: "es-ES"}, {Value: "es", Q: 0.9}},
	{{Value: "ja-JP"}, {Value: "ja", Q: 0.9}},
	{{Value: "ko-KR"}, {Value: "ko", Q: 0.9}},
	{{Value: "pt-BR"}, {Value: "pt", Q: 0.9}},
	{{Value: "ru-RU"}, {Value: "ru", Q: 0.9}},
	{{Value: "tr-TR"}, {Value: "tr", Q: 0.9}},
	{{Value: "it-IT"}, {Value: "it", Q: 0.9}},
	{{Value: "pl-PL"}, {Value: "pl", Q: 0.9}},
	{{Value: "nl-NL"}, {Value: "nl", Q: 0.9}},
	{{Value: "sv-SE"}, {Value: "sv", Q: 0.9}},
	{{Value: "ar-EG"}, {Value: "ar", Q: 0.9}},
	{{Value: "cs-CZ"}, {Value: "cs", Q: 0.9}},
}

func NewGenerator(opts ...Option) *Generator {
	g := &Generator{
		browsers:               []Browser{BrowserRandom},
		platforms:              []Platform{PlatformRandom},
		os:                     []OperatingSystem{OSRandom},
		minVersion:             defaultMinChromiumVersion,
		maxVersion:             defaultMaxChromiumVersion,
		languageProfiles:       defaultLanguages,
		requestType:            RequestTypeNavigate,
		headerSorter:           PriorityHeaderSorter,
		fullFingerprint:        false,
		h2Only:                 true,
		fingerprintProfile:     FingerprintProfileNormal,
		h2RandomizationProfile: H2RandomizationProfileNone,
		useBotAgents:           false,
		botAgentTypes:          nil,
		acceptEncodingEnabled:  false,
		acceptEnabled:          true,
	}

	g.agentPool.New = func() any {
		return new(Agent)
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

func (g *Generator) Generate() (*Agent, error) {
	agent := g.agentPool.Get().(*Agent)
	if agent.Headers == nil {
		agent.Headers = make(http.Header, 16)
	}

	ok := false
	defer func() {
		if !ok {
			g.ReleaseAgent(agent)
		}
	}()

	if g.useBotAgents {
		var eligibleBots []botProfile
		if len(g.botAgentTypes) == 0 {
			eligibleBots = allBotProfiles
		} else {
			for _, botName := range g.botAgentTypes {
				if profiles, botOk := botProfileCategories[botName]; botOk {
					eligibleBots = append(eligibleBots, profiles...)
				}
			}
		}

		if len(eligibleBots) == 0 {
			return nil, fmt.Errorf("%w: %v", ErrNoBotProfiles, g.botAgentTypes)
		}

		chosenProfile := fastrand.Choice(eligibleBots)

		agent.UserAgent = chosenProfile.UserAgent
		agent.ClientHelloID = chosenProfile.HelloID

		for k, v := range chosenProfile.Headers {
			headerSet(agent.Headers, k, v)
		}

		for k, vs := range agent.Headers {
			if len(vs) == 0 {
				delete(agent.Headers, k)
			}
		}

		keysPtr := keysPool.Get().(*[]string)
		keys := (*keysPtr)[:0]
		for k := range agent.Headers {
			keys = append(keys, k)
		}
		PriorityHeaderSorter(keys)
		agent.HeaderOrder = rebuildHeaderOrder(agent.HeaderOrder, keys)
		*keysPtr = keys
		keysPool.Put(keysPtr)

		if g.h2Only {
			agent.H2Settings, agent.h2SettingsPool = GetChromiumH2SettingsWithPool()
		} else {
			agent.H2Settings = nil
			agent.h2SettingsPool = nil
		}

		ok = true
		return agent, nil
	}

	browser, err := g.resolveBrowser()
	if err != nil {
		return nil, err
	}

	profile := browserProfiles[browser]

	chosenPlatform, chosenOS, err := g.resolvePlatformAndOS(browser)
	if err != nil {
		return nil, err
	}

	platformProf := platformProfiles[chosenPlatform]
	osProf := osProfiles[chosenOS]

	finalVersions := g.filterVersions(profile)

	if len(finalVersions) == 0 {
		return nil, fmt.Errorf("%w for %s", ErrNoVersions, browser)
	}

	version := fastrand.Choice(finalVersions)
	versionProf := profile.Versions[version]

	fullVersion := ""
	if profile.ChromiumBased {
		fullVersion = strconv.Itoa(version) + ".0." + strconv.Itoa(versionProf.BuildNumber) + "." + strconv.Itoa(fastrand.IntN(999))
	}

	sb := builderPool.Get().(*strings.Builder)
	defer func() {
		sb.Reset()
		builderPool.Put(sb)
	}()

	componentGenerators := platformProf.ComponentGenerators[profile.Family]
	firstPart := true

	for _, componentGenerator := range componentGenerators {
		part := componentGenerator(profile, osProf, versionProf, fullVersion)
		if part != "" {
			if !firstPart {
				sb.WriteByte(' ')
			}
			sb.WriteString(part)
			firstPart = false
		}
	}

	agent.UserAgent = sb.String()

	headerSorter := g.headerSorter

	if g.fingerprintProfile == FingerprintProfileMaximum {
		headerSorter = ShuffledPriorityHeaderSorter
	}

	if !g.zeroHeader && len(g.languageProfiles) == 0 {
		return nil, ErrNoLanguageProfiles
	}

	if !g.zeroHeader {
		g.buildHeaders(
			agent,
			profile,
			osProf,
			platformProf,
			version,
			fullVersion,
			versionProf,
			headerSorter,
		)
	} else {
		for k, vs := range agent.Headers {
			if len(vs) > 0 {
				agent.Headers[k] = vs[:0]
			}
		}
		agent.HeaderOrder = agent.HeaderOrder[:0]
	}

	if g.h2Only {
		agent.H2Settings, agent.h2SettingsPool = profile.H2SettingsWithPool()
		if g.h2RandomizationProfile != H2RandomizationProfileNone {
			agent.H2Settings = randomizeH2Settings(agent.H2Settings, g.h2RandomizationProfile)
		}
	} else {
		agent.H2Settings = nil
		agent.h2SettingsPool = nil
	}

	if g.fingerprintProfile == FingerprintProfileMaximum {
		agent.ClientHelloSpec = ChromeLatestSpec()
		agent.ClientHelloID = utls.ClientHelloID{}
	} else {
		agent.ClientHelloID = versionProf.TLS.HelloID
		agent.ClientHelloSpec = nil
	}

	ok = true
	return agent, nil
}

func (g *Generator) MustGenerate() *Agent {
	agent, err := g.Generate()
	if err != nil {
		panic(err)
	}
	return agent
}

func (g *Generator) ReleaseAgent(a *Agent) {
	if a == nil {
		return
	}
	resetAgentFields(a)
	g.agentPool.Put(a)
}

func resetAgentFields(a *Agent) {
	a.UserAgent = ""
	for k, vs := range a.Headers {
		if len(vs) > 0 {
			a.Headers[k] = vs[:0]
		}
	}

	a.HeaderOrder = a.HeaderOrder[:0]
	a.ClientHelloSpec = nil
	a.ClientHelloID = utls.ClientHelloID{}
	if a.H2Settings != nil && a.h2SettingsPool != nil {
		releaseH2Settings(a.h2SettingsPool, a.H2Settings)
	}
	a.H2Settings = nil
	a.h2SettingsPool = nil
}

func (g *Generator) resolveBrowser() (Browser, error) {
	var potentialBrowsers []Browser

	if slices.Contains(g.browsers, BrowserRandom) {
		potentialBrowsers = allRealBrowsers
	} else {
		potentialBrowsers = g.browsers
	}

	if len(potentialBrowsers) == 0 {
		return "", ErrNoBrowsers
	}

	return fastrand.Choice(potentialBrowsers), nil
}

type platformOSCombo struct {
	platform Platform
	os       OperatingSystem
}

func (g *Generator) resolvePlatformAndOS(browser Browser) (Platform, OperatingSystem, error) {
	combosPtr := comboPool.Get().(*[]platformOSCombo)
	validCombos := (*combosPtr)[:0]
	defer func() {
		*combosPtr = validCombos
		comboPool.Put(combosPtr)
	}()

	userPlatforms := g.platforms
	if len(userPlatforms) == 1 && userPlatforms[0] == PlatformRandom {
		userPlatforms = allRealPlatforms
	}

	userOSes := g.os
	if len(userOSes) == 1 && userOSes[0] == OSRandom {
		userOSes = allRealOS
	}

	for _, p := range userPlatforms {
		for _, o := range userOSes {
			var concreteOSBuf [1]OperatingSystem
			var concreteOSes []OperatingSystem
			if o == OSMac {
				concreteOSes = macArchitectures
			} else {
				concreteOSBuf[0] = o
				concreteOSes = concreteOSBuf[:]
			}

			for _, concreteOS := range concreteOSes {
				osProfile, osProfileExists := osProfiles[concreteOS]
				if !osProfileExists {
					continue
				}

				if (p == PlatformMobile) != osProfile.IsMobile {
					continue
				}

				isValidForBrowser := false
				switch browser {
				case BrowserSafari:
					if (p == PlatformMobile && concreteOS == OSiOS) || (p == PlatformDesktop && osProfile.IsMacOS) {
						isValidForBrowser = true
					}
				case BrowserFirefox:
					if (p == PlatformMobile && concreteOS == OSAndroid) || (p == PlatformDesktop && concreteOS != OSiOS && concreteOS != OSAndroid) {
						isValidForBrowser = true
					}
				default:
					isValidForBrowser = true
				}

				if isValidForBrowser {
					validCombos = append(validCombos, platformOSCombo{p, concreteOS})
				}
			}
		}
	}

	if len(validCombos) == 0 {
		return "", "", fmt.Errorf("%w for browser %s", ErrNoPlatformOSCombo, browser)
	}

	chosenCombo := fastrand.Choice(validCombos)

	return chosenCombo.platform, chosenCombo.os, nil
}

func (g *Generator) buildHeaders(agent *Agent, browser browserProfile, os osProfile, platform platformProfile, version int, fullVersion string, versionProf versionProfile, sorter HeaderSorter) {
	header := agent.Headers

	var acceptTemplate [][]AcceptHeaderPart
	if g.requestType == RequestTypeXHR {
		acceptTemplate = versionProf.AcceptHeaderPatternsXHR
	} else {
		acceptTemplate = versionProf.AcceptHeaderPatterns
	}

	languageTemplate := fastrand.Choice(g.languageProfiles)

	if g.acceptEnabled {
		headerSet(header, hAccept, buildAcceptHeader(fastrand.Choice(acceptTemplate)))
	}
	if g.acceptEncodingEnabled {
		headerSet(header, hAcceptEncoding, generateAcceptEncoding())
	}

	headerSet(header, hAcceptLanguage, buildAcceptHeader(languageTemplate))

	if browser.ChromiumBased {
		headerSet(header, hSecChUa, buildSecChUa(browser.Brand, strconv.Itoa(version), false, true))
		headerSet(header, hSecChUaMobile, platform.MobileHint)
		headerSet(header, hSecChUaPlatform, os.PlatformQuote)

		if g.fullFingerprint {
			headerSet(header, hSecChUaFullVersionList, buildSecChUa(browser.Brand, fullVersion, true, true))
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
	}

	if g.fingerprintProfile == FingerprintProfileExtreme {
		for k := range header {
			if strings.HasPrefix(k, "sec-") && fastrand.Bool() {
				header[k] = header[k][:0]
			}
		}
	}

	if g.requestType == RequestTypeNavigate && browser.Brand == "Brave" {
		headerSet(header, hSecGpc, "1")
	}

	switch g.requestType {
	case RequestTypeNavigate:
		headerSet(header, hSecFetchDest, "document")
		headerSet(header, hSecFetchMode, "navigate")
		headerSet(header, hSecFetchSite, "none")
		headerSet(header, hSecFetchUser, "?1")
		headerSet(header, hUpgradeInsecureRequests, "1")
	case RequestTypeSubresource:
		headerSet(header, hSecFetchDest, fastrand.Choice(subresourceDests))
		headerSet(header, hSecFetchMode, "no-cors")
		headerSet(header, hSecFetchSite, "same-origin")
	case RequestTypeXHR:
		headerSet(header, hSecFetchDest, "empty")
		headerSet(header, hSecFetchMode, "cors")
		headerSet(header, hSecFetchSite, "same-origin")
	}

	keysPtr := keysPool.Get().(*[]string)
	keys := (*keysPtr)[:0]
	for k, vs := range header {
		if len(vs) == 0 {
			continue
		}
		keys = append(keys, k)
	}

	sorter(keys)
	agent.HeaderOrder = rebuildHeaderOrder(agent.HeaderOrder, keys)

	for k, vs := range header {
		if len(vs) == 0 {
			delete(header, k)
		}
	}

	*keysPtr = keys
	keysPool.Put(keysPtr)
}

func buildSecChUa(brand, version string, isFull, randomize bool) string {
	var grease greaseBrandParts
	if randomize {
		grease = fastrand.Choice(greaseBrandParsed)
	} else {
		grease = greaseBrandDefault
	}

	v := version
	if !isFull {
		if idx := strings.IndexByte(version, '.'); idx >= 0 {
			v = version[:idx]
		}
	}

	greaseVersion := grease.Version
	if isFull {
		greaseVersion = `"99.0.0.0"`
	}

	sb := builderPool.Get().(*strings.Builder)
	defer func() {
		sb.Reset()
		builderPool.Put(sb)
	}()

	order := [3]int{0, 1, 2}
	if randomize {
		fastrand.Shuffle(3, func(i, j int) { order[i], order[j] = order[j], order[i] })
	}

	for i := 0; i < 3; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		switch order[i] {
		case 0:
			sb.WriteString(`"Chromium";v="`)
			sb.WriteString(v)
			sb.WriteString(`"`)
		case 1:
			sb.WriteString(`"`)
			sb.WriteString(brand)
			sb.WriteString(`";v="`)
			sb.WriteString(v)
			sb.WriteString(`"`)
		case 2:
			sb.WriteString(grease.Key)
			sb.WriteString(`;v=`)
			sb.WriteString(greaseVersion)
		}
	}
	return sb.String()
}

var acceptEncodingEncodings = [3]string{"gzip", "deflate", "br"}

func generateAcceptEncoding() string {
	var encodings [3]string
	encodings = acceptEncodingEncodings

	fastrand.Shuffle(3, func(i, j int) {
		encodings[i], encodings[j] = encodings[j], encodings[i]
	})

	sb := builderPool.Get().(*strings.Builder)
	defer func() {
		sb.Reset()
		builderPool.Put(sb)
	}()

	for i := 0; i < 3; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(encodings[i])
	}
	if fastrand.IntN(2) == 1 {
		sb.WriteString(", zstd")
	}
	return sb.String()
}

func buildAcceptHeader(parts []AcceptHeaderPart) string {
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

		if len(part.Extras) > 0 {
			for _, extra := range part.Extras {
				sb.WriteByte(';')
				sb.WriteString(extra)
			}
		}

		if part.Q > 0 {
			q := part.Q - fastrand.Float64()*0.05
			if q < 0.05 {
				q = 0.05
			}
			sb.WriteString(";q=")
			appendQValue(sb, q)
		}
	}
	return sb.String()
}

func appendQValue(sb *strings.Builder, q float64) {
	if q >= 1.0 {
		sb.WriteString("1.0")
		return
	}
	tenths := int(q*10 + 0.5)
	if tenths >= 10 {
		sb.WriteString("1.0")
		return
	}
	sb.WriteString("0.")
	sb.WriteByte('0' + byte(tenths))
}

func (g *Generator) filterVersions(profile browserProfile) []int {
	defaultRange := g.minVersion == defaultMinChromiumVersion && g.maxVersion == defaultMaxChromiumVersion

	if defaultRange && g.h2Only {
		return profile.H2VersionKeys
	}

	if defaultRange && !g.h2Only {
		return profile.VersionKeys
	}

	allVersions := profile.VersionKeys
	possibleVersions := make([]int, 0, len(allVersions))
	for _, v := range allVersions {
		if v < g.minVersion || v > g.maxVersion {
			continue
		}
		if g.h2Only && !profile.Versions[v].SupportsH2 {
			continue
		}
		possibleVersions = append(possibleVersions, v)
	}
	return possibleVersions
}

func rebuildHeaderOrder(existing, keys []string) []string {
	needed := len(pseudoHeaders) + len(keys)
	if cap(existing) >= needed {
		existing = existing[:needed]
	} else {
		existing = make([]string, needed)
	}
	n := copy(existing, pseudoHeaders)
	copy(existing[n:], keys)
	return existing
}
