package legitagent

import "testing"

func BenchmarkGenerateChrome(b *testing.B) {
	g := NewGenerator(WithBrowsers(BrowserChrome), WithOS(OSWindows11), WithPlatforms(PlatformDesktop))
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		agent, err := g.Generate()
		if err != nil {
			b.Fatal(err)
		}
		g.ReleaseAgent(agent)
	}
}

func BenchmarkGenerateRandom(b *testing.B) {
	g := NewGenerator()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		agent, err := g.Generate()
		if err != nil {
			b.Fatal(err)
		}
		g.ReleaseAgent(agent)
	}
}

func BenchmarkGenerateFullFingerprint(b *testing.B) {
	g := NewGenerator(
		WithBrowsers(BrowserChrome),
		WithOS(OSWindows11),
		WithFullFingerprint(true),
		WithAcceptEncoding(true),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		agent, err := g.Generate()
		if err != nil {
			b.Fatal(err)
		}
		g.ReleaseAgent(agent)
	}
}

func BenchmarkGenerateMaximum(b *testing.B) {
	g := NewGenerator(
		WithBrowsers(BrowserChrome),
		WithFingerprintProfile(FingerprintProfileMaximum),
		WithH2Randomization(H2RandomizationProfileMaximum),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		agent, err := g.Generate()
		if err != nil {
			b.Fatal(err)
		}
		g.ReleaseAgent(agent)
	}
}

func BenchmarkGenerateBot(b *testing.B) {
	g := NewGenerator(WithBotAgents())
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		agent, err := g.Generate()
		if err != nil {
			b.Fatal(err)
		}
		g.ReleaseAgent(agent)
	}
}

func BenchmarkFromUserAgentString(b *testing.B) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36"
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, err := FromUserAgentString(ua, RequestTypeNavigate)
		if err != nil {
			b.Fatal(err)
		}
	}
}
