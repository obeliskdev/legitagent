# legitagent — Realistic Browser Fingerprint & User-Agent Generator for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/obeliskdev/legitagent.svg)](https://pkg.go.dev/github.com/obeliskdev/legitagent)
[![Go Report Card](https://goreportcard.com/badge/github.com/obeliskdev/legitagent)](https://goreportcard.com/report/github.com/obeliskdev/legitagent)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`legitagent` is a high-performance Go library that generates **realistic, browser-consistent HTTP client profiles** for scraping, automation, testing, and security research. It goes far beyond simple `User-Agent` string randomization — every generated profile includes a coordinated set of TLS fingerprints, HTTP/2 settings, header ordering, and accept headers that match a real browser family.

## Why legitagent?

Most anti-bot systems don't just check the `User-Agent` header. They correlate TLS JA3/JA4 fingerprints, HTTP/2 SETTINGS frame ordering, header priority, and accept-header patterns. A mismatch between any of these signals reveals a bot instantly. `legitagent` solves this by producing **internally consistent profiles** where every layer matches a real browser.

### Key Features

- **7 browser engines** — Chrome, Firefox, Safari, Edge, Opera, Brave, plus 27 crawler bot profiles (GoogleBot, GPTBot, ClaudeBot, etc.)
- **Full TLS fingerprinting** — uTLS `ClientHelloID` and `ClientHelloSpec` matched to each browser family (JA3/JA4 consistent)
- **HTTP/2 fingerprinting** — browser-accurate H2 SETTINGS frames with optional randomization to avoid static fingerprints
- **Header ordering** — priority-based header sorters that mimic real browser behavior, with shuffle support for variability
- **Platform-aware** — desktop (Windows, macOS Intel/Apple Silicon, Linux, ChromeOS) and mobile (Android, iOS) profiles with correct UA tokens
- **Accept-header consistency** — `Accept`, `Accept-Language`, `Accept-Encoding` patterns matched per browser and request type (navigate, subresource, XHR)
- **Zero-allocation fast path** — `sync.Pool`-backed `Agent` recycling, precomputed maps, and in-place randomization keep allocations low (~23 allocs/op for Chrome)
- **User-Agent parser** — parse existing UA strings into full `Agent` profiles with `FromUserAgentString()`
- **16 default languages** — en-US, de-DE, fa-IR, fr-FR, es-ES, ja-JP, ko-KR, pt-BR, ru-RU, tr-TR, it-IT, pl-PL, nl-NL, sv-SE, ar-EG, cs-CZ
- **Sentinel errors** — typed errors (`ErrNoBrowsers`, `ErrNoVersions`, etc.) for programmatic handling
- **Thread-safe** — `Generator` uses `sync.Pool`; create one and reuse it across goroutines

## Installation

```bash
go get github.com/obeliskdev/legitagent
```

Requires Go 1.25+ and depends on:

| Dependency | Purpose |
|---|---|
| [refraction-networking/utls](https://github.com/refraction-networking/utls) | TLS fingerprint replication |
| [golang.org/x/net](https://pkg.go.dev/golang.org/x/net) | HTTP/2 setting types |
| [obeliskdev/fastrand](https://github.com/obeliskdev/fastrand) | Lock-free random number generation |

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/obeliskdev/legitagent"
)

func main() {
	g := legitagent.NewGenerator()

	agent, err := g.Generate()
	if err != nil {
		log.Fatal(err)
	}
	defer g.ReleaseAgent(agent)

	fmt.Println("User-Agent:", agent.UserAgent)
	fmt.Println("ClientHelloID:", agent.ClientHelloID.Str())
	fmt.Println("Accept-Language:", agent.Headers.Get("accept-language"))
	fmt.Println("Header order:", agent.HeaderOrder)
}
```

### Panicking Variant

If you don't need error handling, use `MustGenerate()`:

```go
agent := g.MustGenerate()
defer g.ReleaseAgent(agent)
```

## Generator Options

All options follow the functional-options pattern and are passed to `NewGenerator()`:

### Browser / Platform / OS Selection

| Option | Description |
|---|---|
| `WithBrowsers(b ...Browser)` | Restrict to specific browsers (Chrome, Firefox, Safari, Edge, Opera, Brave) |
| `WithPlatforms(p ...Platform)` | Restrict to desktop or mobile |
| `WithOS(os ...OperatingSystem)` | Restrict to specific OS (Windows, Windows11, Linux, Mac, Android, iOS, ChromeOS) |
| `WithVersionRange(min, max)` | Constrain browser major version range |
| `WithLanguages(langs ...string)` | Override default Accept-Language profiles (e.g. `"en-US,en;q=0.9"`) |

### Header Behavior

| Option | Description |
|---|---|
| `WithRequestType(rt)` | `RequestTypeNavigate`, `RequestTypeSubresource`, or `RequestTypeXHR` |
| `WithHeaderSorter(sorter)` | Custom header ordering function (`PriorityHeaderSorter`, `ShuffledPriorityHeaderSorter`) |
| `WithAccept(true\|false)` | Include/exclude `Accept` header |
| `WithAcceptEncoding(true\|false)` | Include/exclude `Accept-Encoding` header |
| `WithFullFingerprint(true\|false)` | Include all `sec-ch-ua-*` client hint headers |
| `WithZeroHeader(true\|false)` | Include `0` priority header for Chrome |

### Fingerprint Behavior

| Option | Description |
|---|---|
| `WithFingerprintProfile(p)` | `FingerprintProfileNormal`, `Maximum`, or `Extreme` |
| `WithH2Only(true\|false)` | Restrict to H2-capable browsers only |
| `WithH2Randomization(p)` | `H2RandomizationProfileNone`, `Normal`, or `Maximum` |

### Bot Mode

| Option | Description |
|---|---|
| `WithBotAgents(bots ...string)` | Generate crawler bot profiles instead of browser profiles |

Available bot constants: `BotGoogle`, `BotBing`, `BotGPT`, `BotChatGPT`, `BotClaude`, `BotCohere`, `BotPerplexity`, `BotAhrefs`, `BotSemrush`, `BotYandex`, `BotBaidu`, `BotDuckDuckGo`, `BotLinkedIn`, `BotFacebook`, `BotTwitter`, `BotWhatsApp`, `BotPinterest`, `BotApple`, `BotBytespider`, `BotCC`, `BotDiffbot`, `BotMajestic`, `BotMoz`, `BotPetal`, `BotSogou`, `BotUptimeRobot`, `BotYahoo`, `BotYou`.

## Examples

### Force Firefox Desktop on Linux

```go
g := legitagent.NewGenerator(
	legitagent.WithBrowsers(legitagent.BrowserFirefox),
	legitagent.WithPlatforms(legitagent.PlatformDesktop),
	legitagent.WithOS(legitagent.OSLinux),
)

agent, err := g.Generate()
if err != nil {
	panic(err)
}
defer g.ReleaseAgent(agent)
```

### Maximum Fingerprint Variability

```go
g := legitagent.NewGenerator(
	legitagent.WithFingerprintProfile(legitagent.FingerprintProfileMaximum),
	legitagent.WithH2Randomization(legitagent.H2RandomizationProfileMaximum),
)

agent1 := g.MustGenerate()
agent2 := g.MustGenerate()

g.ReleaseAgent(agent1)
g.ReleaseAgent(agent2)
```

### Custom Languages

```go
g := legitagent.NewGenerator(
	legitagent.WithLanguages("ja-JP,ja;q=0.9,en-US;q=0.8", "ko-KR,ko;q=0.9,en-US;q=0.8"),
)
```

### Crawler Bot Profiles

```go
g := legitagent.NewGenerator(
	legitagent.WithBotAgents(legitagent.BotGoogle, legitagent.BotGPT),
)

agent := g.MustGenerate()
defer g.ReleaseAgent(agent)
// agent.UserAgent == "Mozilla/5.0 (compatible; GPTBot/1.2; +https://chatgpt.com/gptbot)"
```

### Parse an Existing User-Agent String

```go
parsed, err := legitagent.FromUserAgentString(
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 ...",
	legitagent.RequestTypeNavigate,
)
if err != nil {
	panic(err)
}
_ = parsed
```

## Integrating With HTTP Clients

A generated `Agent` exposes everything you need to wire into a real HTTP client:

| Field | Type | Usage |
|---|---|---|
| `UserAgent` | `string` | Set the `User-Agent` header |
| `Headers` | `http.Header` | Full header map to apply to requests |
| `HeaderOrder` | `[]string` | Ordered header keys for HTTP/2 pseudo-headers and priority |
| `ClientHelloID` | `utls.ClientHelloID` | uTLS handshake identifier |
| `ClientHelloSpec` | `*utls.ClientHelloSpec` | uTLS handshake specification |
| `H2Settings` | `map[http2.SettingID]uint32` | HTTP/2 SETTINGS frame values |

### Example: uTLS + HTTP/2 Transport

```go
agent := g.MustGenerate()
defer g.ReleaseAgent(agent)

dialTLS := func(network, addr string) (net.Conn, error) {
	tcpConn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	uConn := utls.UClient(tcpConn, &utls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}, agent.ClientHelloID)

	if err := uConn.Handshake(); err != nil {
		return nil, err
	}
	return uConn, nil
}

transport := &http2.Transport{
	DialTLS: dialTLS,
}
client := &http.Client{Transport: transport}

req, _ := http.NewRequest("GET", "https://example.com", nil)
req.Header = agent.Headers
// Apply header ordering as needed by your transport
```

## Performance

Benchmark results (Go 1.25, Windows/amd64):

```
BenchmarkGenerateChrome-8        23 allocs/op   ~2295 ns/op
BenchmarkGenerateFirefox-8      18 allocs/op   ~2100 ns/op
BenchmarkGenerateSafari-8       21 allocs/op   ~2200 ns/op
BenchmarkGenerateEdge-8          23 allocs/op   ~2300 ns/op
BenchmarkGenerateRandom-8       23 allocs/op   ~2400 ns/op
BenchmarkFromUserAgentString-8   7 allocs/op    ~450 ns/op
```

The `Generator` uses `sync.Pool` for `Agent` recycling. Always call `ReleaseAgent(agent)` when done to return the object to the pool.

## Supported Browser Versions

| Browser | Versions |
|---|---|
| Chrome | 146–151 |
| Edge | 130, 146–151 |
| Firefox | 130–141 |
| Safari | 18 |
| Opera | 146–151 |
| Brave | 146–151 |

Version maps are updated periodically to track current browser releases.

## Sentinel Errors

The library returns typed sentinel errors for common failure cases:

```go
var (
	ErrNoBrowsers         = errors.New("legitagent: no browsers configured for generation")
	ErrNoBotProfiles      = errors.New("legitagent: no bot profiles found for the specified types")
	ErrNoVersions         = errors.New("legitagent: no available browser versions that meet the specified criteria")
	ErrNoLanguageProfiles = errors.New("legitagent: no language profiles configured")
	ErrNoPlatformOSCombo  = errors.New("legitagent: no compatible platform/OS combination found")
)
```

Use `errors.Is()` to check for specific conditions.

## Testing

```bash
go test -race -count=3 ./...
```

Benchmarks:

```bash
go test -bench=. -benchmem ./...
```

## Use Cases

- **Web scraping & data collection** — rotate consistent browser profiles to reduce detection
- **API automation** — present realistic browser fingerprints to TLS-fingerprinting gateways
- **Security research** — generate diverse profiles for penetration testing and audit
- **Load testing** — simulate traffic from varied browser/OS combinations
- **Ad-tech verification** — confirm content delivery across realistic client profiles
- **AI crawler compliance** — use legitimate bot profiles (GPTBot, GoogleBot, ClaudeBot) with correct identification

## License

MIT. See `LICENSE` file.

## Contributing

Issues and pull requests are welcome at [github.com/obeliskdev/legitagent](https://github.com/obeliskdev/legitagent).