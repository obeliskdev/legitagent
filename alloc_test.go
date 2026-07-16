package legitagent

import (
	"testing"
)

func TestAllocsGenerate(t *testing.T) {
	g := NewGenerator()

	allocs := testing.AllocsPerRun(100, func() {
		agent, err := g.Generate()
		if err != nil {
			t.Fatal(err)
		}
		g.ReleaseAgent(agent)
	})

	if allocs > 30 {
		t.Errorf("Generate allocated %v times, expected <= 30", allocs)
	}
}

func TestAllocsGenerateReuse(t *testing.T) {
	g := NewGenerator()

	agent, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}
	g.ReleaseAgent(agent)

	allocs := testing.AllocsPerRun(100, func() {
		agent, err := g.Generate()
		if err != nil {
			t.Fatal(err)
		}
		g.ReleaseAgent(agent)
	})

	if allocs > 25 {
		t.Errorf("Generate (pooled) allocated %v times, expected <= 25", allocs)
	}
}

func TestAllocsH2SettingsPool(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		m := GetChromiumH2Settings()
		ReleaseChromiumH2Settings(m)
	})

	if allocs > 1 {
		t.Errorf("GetChromiumH2Settings+Release allocated %v times, expected <= 1", allocs)
	}
}

func TestAllocsFromUserAgentStringReuse(t *testing.T) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"

	agent, err := FromUserAgentString(ua, RequestTypeNavigate)
	if err != nil {
		t.Fatal(err)
	}
	ReleaseParserAgent(agent)

	allocs := testing.AllocsPerRun(100, func() {
		agent, err := FromUserAgentString(ua, RequestTypeNavigate)
		if err != nil {
			t.Fatal(err)
		}
		ReleaseParserAgent(agent)
	})

	if allocs > 40 {
		t.Errorf("FromUserAgentString (pooled) allocated %v times, expected <= 40", allocs)
	}
}

func TestAllocsShuffledPriorityHeaderSorter(t *testing.T) {
	keys := []string{
		"host", "connection", "user-agent", "accept", "accept-encoding",
		"accept-language", "sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform",
		"sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest",
		"upgrade-insecure-requests",
	}

	allocs := testing.AllocsPerRun(100, func() {
		k := make([]string, len(keys))
		copy(k, keys)
		ShuffledPriorityHeaderSorter(k)
	})

	if allocs > 4 {
		t.Errorf("ShuffledPriorityHeaderSorter allocated %v times, expected <= 4", allocs)
	}
}

func TestAllocsBuildSecChUa(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		_ = buildSecChUa("Google Chrome", "145.0.7632.0", false, true)
	})

	if allocs > 5 {
		t.Errorf("buildSecChUa allocated %v times, expected <= 5", allocs)
	}
}

func TestAllocsGenerateAcceptEncoding(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		_ = generateAcceptEncoding()
	})

	if allocs > 3 {
		t.Errorf("generateAcceptEncoding allocated %v times, expected <= 3", allocs)
	}
}
