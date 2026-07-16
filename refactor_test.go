package legitagent

import (
	"strings"
	"testing"
)

func TestDefaultLanguages_PackageLevel(t *testing.T) {
	g1 := NewGenerator()
	g2 := NewGenerator()

	if len(g1.languageProfiles) != len(g2.languageProfiles) {
		t.Error("both generators should share the same defaultLanguages slice")
	}
	if len(defaultLanguages) != 16 {
		t.Errorf("defaultLanguages should have 16 entries, got %d", len(defaultLanguages))
	}
}

func TestFilterVersions_CombinedPass(t *testing.T) {
	g := NewGenerator(
		WithBrowsers(BrowserChrome),
		WithVersionRange(120, 133),
		WithH2Only(true),
	)
	profile := browserProfiles[BrowserChrome]
	result := g.filterVersions(profile)

	if len(result) == 0 {
		t.Fatal("should find at least one matching version")
	}
	for _, v := range result {
		if v < 120 {
			t.Errorf("version %d should be >= minVersion 120", v)
		}
		if v > 133 {
			t.Errorf("version %d should be <= maxVersion 133", v)
		}
		if !profile.Versions[v].SupportsH2 {
			t.Errorf("version %d should support H2", v)
		}
	}
}

func TestFilterVersions_CombinedPass_NoH2(t *testing.T) {
	g := NewGenerator(
		WithBrowsers(BrowserChrome),
		WithVersionRange(120, 133),
		WithH2Only(false),
	)
	profile := browserProfiles[BrowserChrome]
	result := g.filterVersions(profile)

	if len(result) == 0 {
		t.Fatal("should find at least one matching version")
	}
	for _, v := range result {
		if v < 120 || v > 133 {
			t.Errorf("version %d should be in [120, 133]", v)
		}
	}
}

func TestBrowserSuffixGenerator_NoSplit(t *testing.T) {
	bp := browserProfiles[BrowserOpera]
	result := BrowserSuffixGenerator(bp, osProfile{}, versionProfile{}, "145.0.7632.100")
	if result != "OPR/145" {
		t.Errorf("expected 'OPR/145', got %q", result)
	}
}

func TestBrowserSuffixGenerator_EmptySuffix(t *testing.T) {
	bp := browserProfiles[BrowserChrome]
	result := BrowserSuffixGenerator(bp, osProfile{}, versionProfile{}, "145.0.7632.100")
	if result != "" {
		t.Errorf("Chrome has empty UASuffix, expected empty string, got %q", result)
	}
}

func TestResolveDeviceToken_Android(t *testing.T) {
	op := osProfiles[OSAndroid]
	token := resolveDeviceToken(op)
	if strings.Contains(token, "{device_model}") {
		t.Error("device model placeholder should be replaced")
	}
	if !strings.Contains(token, "Linux; Android 15;") {
		t.Errorf("token should contain Android platform prefix, got %q", token)
	}
}

func TestResolveDeviceToken_NonAndroid(t *testing.T) {
	op := osProfiles[OSWindows11]
	token := resolveDeviceToken(op)
	if token != op.PlatformToken {
		t.Errorf("non-Android token should be returned as-is, got %q", token)
	}
}

func TestResolveDeviceToken_Idempotent(t *testing.T) {
	op := osProfiles[OSLinux]
	t1 := resolveDeviceToken(op)
	t2 := resolveDeviceToken(op)
	if t1 != t2 {
		t.Error("non-Android tokens should be identical across calls")
	}
}

func TestNewGenerator_DefaultLanguagesInitialized(t *testing.T) {
	g := NewGenerator()
	if len(g.languageProfiles) == 0 {
		t.Error("language profiles should be initialized")
	}
}

func TestResetAgentFields_ClearsAll(t *testing.T) {
	g := NewGenerator()
	agent, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}

	resetAgentFields(agent)

	if agent.UserAgent != "" {
		t.Error("UserAgent should be cleared")
	}
	if agent.ClientHelloSpec != nil {
		t.Error("ClientHelloSpec should be nil")
	}
	if agent.H2Settings != nil {
		t.Error("H2Settings should be nil")
	}
	if agent.h2SettingsPool != nil {
		t.Error("h2SettingsPool should be nil")
	}
	if len(agent.HeaderOrder) != 0 {
		t.Error("HeaderOrder should be empty")
	}
}

func TestReleaseAgent_PutsBackInPool(t *testing.T) {
	g := NewGenerator()
	agent, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}
	g.ReleaseAgent(agent)

	agent2, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if agent2 != agent {
		t.Error("second Generate should reuse pooled agent")
	}
	g.ReleaseAgent(agent2)
}

func TestDetectOSFromUA_AllPlatforms(t *testing.T) {
	tests := []struct {
		ua      string
		browser Browser
		want    OperatingSystem
	}{
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/145.0.0.0", BrowserChrome, OSWindows11},
		{"Mozilla/5.0 (Linux; Android 15; Pixel 8) Chrome/145.0.0.0", BrowserChrome, OSAndroid},
		{"Mozilla/5.0 (X11; CrOS x86_64 15923.0.0) Chrome/145.0.0.0", BrowserChrome, OSChromeOS},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) Safari/604.1", BrowserSafari, OSiOS},
		{"Mozilla/5.0 (iPad; CPU OS 18_5 like Mac OS X) Safari/604.1", BrowserSafari, OSiOS},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/605.1.15", BrowserSafari, osMacIntel},
		{"Mozilla/5.0 (X11; Linux x86_64) Firefox/140.0", BrowserFirefox, OSLinux},
	}

	for _, tt := range tests {
		got := detectOSFromUA(tt.ua, tt.browser)
		if got != tt.want {
			t.Errorf("detectOSFromUA(%q) = %s, want %s", tt.ua, got, tt.want)
		}
	}
}

func TestDetectOSFromUA_SafariMobileRedirectsToIOS(t *testing.T) {
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Mobile/15E148 Safari/604.1"
	got := detectOSFromUA(ua, BrowserSafari)
	if got != OSiOS {
		t.Errorf("Safari with Mobile/ should be iOS, got %s", got)
	}
}

func TestBuildSecChUa_ContainsBrandAndChromium(t *testing.T) {
	for i := 0; i < 50; i++ {
		result := buildSecChUa("Google Chrome", "145.0.7632.0", false, true)
		if !strings.Contains(result, `"Chromium"`) {
			t.Errorf("buildSecChUa should contain Chromium, got %q", result)
		}
		if !strings.Contains(result, `"Google Chrome"`) {
			t.Errorf("buildSecChUa should contain brand, got %q", result)
		}
		if !strings.Contains(result, `v="145"`) {
			t.Errorf("buildSecChUa should contain major version 145, got %q", result)
		}
	}
}

func TestBuildSecChUa_FullVersion(t *testing.T) {
	result := buildSecChUa("Google Chrome", "145.0.7632.100", true, false)
	if !strings.Contains(result, `v="145.0.7632.100"`) {
		t.Errorf("full version should contain full version string, got %q", result)
	}
	if !strings.Contains(result, `"99.0.0.0"`) {
		t.Errorf("full version should use grease 99.0.0.0, got %q", result)
	}
}

func TestBuildAcceptHeaderStatic_Format(t *testing.T) {
	parts := []AcceptHeaderPart{
		{Value: "text/html"},
		{Value: "application/xml", Q: 0.9},
		{Value: "*/*", Q: 0.8, Extras: []string{"x=1"}},
	}
	result := buildAcceptHeaderStatic(parts)
	if !strings.Contains(result, "text/html") {
		t.Errorf("should contain text/html, got %q", result)
	}
	if !strings.Contains(result, "application/xml;q=0.9") {
		t.Errorf("should contain Q value, got %q", result)
	}
	if !strings.Contains(result, "*/*;x=1;q=0.8") {
		t.Errorf("should contain extras and Q, got %q", result)
	}
}

func TestGenerateAcceptEncoding_Format(t *testing.T) {
	for i := 0; i < 100; i++ {
		result := generateAcceptEncoding()
		if !strings.Contains(result, "gzip") {
			t.Errorf("should contain gzip, got %q", result)
		}
		if !strings.Contains(result, "deflate") {
			t.Errorf("should contain deflate, got %q", result)
		}
		if !strings.Contains(result, "br") {
			t.Errorf("should contain br, got %q", result)
		}
		commaCount := strings.Count(result, ", ")
		if commaCount < 2 {
			t.Errorf("should have at least 2 separators, got %d in %q", commaCount, result)
		}
	}
}

func TestPriorityHeaderSorter_UnknownHeadersLast(t *testing.T) {
	keys := []string{"x-custom-header", "host", "accept", "x-another-unknown"}
	PriorityHeaderSorter(keys)
	if keys[0] != "host" {
		t.Errorf("host should be first, got %s", keys[0])
	}
	if keys[1] != "accept" {
		t.Errorf("accept should be second, got %s", keys[1])
	}
}

func TestShuffledPriorityHeaderSorter_PreservesPriorityGroups(t *testing.T) {
	keys := []string{
		"accept", "host", "user-agent", "accept-language",
		"x-custom-1", "x-custom-2",
	}
	ShuffledPriorityHeaderSorter(keys)
	if keys[0] != "host" {
		t.Errorf("host (priority 10) should be first, got %s", keys[0])
	}
}

func TestOSProfile_IsMacOS(t *testing.T) {
	for _, os := range []OperatingSystem{osMacIntel, osMacAppleSilicon} {
		op := osProfiles[os]
		if !op.IsMacOS {
			t.Errorf("%s should have IsMacOS=true", os)
		}
	}
	for _, os := range []OperatingSystem{OSWindows11, OSLinux, OSAndroid, OSiOS} {
		op := osProfiles[os]
		if op.IsMacOS {
			t.Errorf("%s should have IsMacOS=false", os)
		}
	}
}

func TestAppendQValue(t *testing.T) {
	tests := []struct {
		q    float64
		want string
	}{
		{1.0, "1.0"},
		{0.9, "0.9"},
		{0.8, "0.8"},
		{0.7, "0.7"},
		{0.5, "0.5"},
		{0.05, "0.1"},
		{0.85, "0.9"},
	}
	for _, tt := range tests {
		var sb strings.Builder
		appendQValue(&sb, tt.q)
		got := sb.String()
		if got != tt.want {
			t.Errorf("appendQValue(%v) = %q, want %q", tt.q, got, tt.want)
		}
	}
}

func TestH2SettingSlice_Precomputed(t *testing.T) {
	if len(ChromiumH2SettingSlice()) == 0 {
		t.Error("ChromiumH2SettingSlice should not be empty")
	}
	if len(GeckoH2SettingSlice()) == 0 {
		t.Error("GeckoH2SettingSlice should not be empty")
	}
	if len(WebKitH2SettingSlice()) == 0 {
		t.Error("WebKitH2SettingSlice should not be empty")
	}

	chromium := ChromiumH2SettingSlice()
	chromiumMap := make(map[uint32]uint32, len(chromium))
	for _, s := range chromium {
		chromiumMap[uint32(s.ID)] = s.Val
	}
	if chromiumMap[uint32(0x1)] != 65536 {
		t.Errorf("chromium HEADER_TABLE_SIZE should be 65536, got %d", chromiumMap[uint32(0x1)])
	}
}

func TestBuildSecChUa_NoIntermediateConcat(t *testing.T) {
	for i := 0; i < 50; i++ {
		result := buildSecChUa("Microsoft Edge", "133.0.2988.0", false, true)
		if !strings.Contains(result, `"Chromium";v="133"`) {
			t.Errorf("should contain Chromium v=133, got %q", result)
		}
		if !strings.Contains(result, `"Microsoft Edge";v="133"`) {
			t.Errorf("should contain brand v=133, got %q", result)
		}
	}
}
