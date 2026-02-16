# legitagent

`legitagent` generates realistic browser agent profiles for Go clients.

It does more than randomize `User-Agent`. A generated profile can include:
- Browser-consistent HTTP headers.
- Header ordering strategy.
- TLS fingerprint material (`ClientHelloID` / `ClientHelloSpec`).
- Browser-style HTTP/2 settings.

This makes it useful when you need request profiles that behave like real browser families.

## Installation

```bash
go get github.com/obeliskdev/legitagent
```

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
	fmt.Println("ClientHelloID:", agent.ClientHelloID)
	fmt.Println("Accept-Language:", agent.Headers.Get("accept-language"))
}
```

## Generator Options

Use options to constrain output and profile behavior:

- Browser/platform/OS selection:
  - `WithBrowsers(...)`
  - `WithPlatforms(...)`
  - `WithOS(...)`
  - `WithVersionRange(min, max)`
- Header behavior:
  - `WithRequestType(...)`
  - `WithHeaderSorter(...)`
  - `WithAccept(true|false)`
  - `WithAcceptEncoding(true|false)`
  - `WithFullFingerprint(true|false)`
  - `WithZeroHeader(true|false)`
- Fingerprint behavior:
  - `WithFingerprintProfile(...)`
  - `WithH2Only(true|false)`
  - `WithH2Randomization(...)`
- Bot mode:
  - `WithBotAgents(...)`

## Example: Force Firefox Desktop Linux

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

## Example: Maximum Fingerprint Variability

```go
g := legitagent.NewGenerator(
	legitagent.WithFingerprintProfile(legitagent.FingerprintProfileMaximum),
	legitagent.WithH2Randomization(legitagent.H2RandomizationProfileMaximum),
)

agent1, _ := g.Generate()
agent2, _ := g.Generate()

g.ReleaseAgent(agent1)
g.ReleaseAgent(agent2)
```

## Parsing Existing User-Agents

You can parse an existing UA string into an `Agent` profile:

```go
parsed, err := legitagent.FromUserAgentString(
	"Mozilla/5.0 ...",
	legitagent.RequestTypeNavigate,
)
if err != nil {
	panic(err)
}
_ = parsed
```

## Integrating With HTTP Clients

A generated `Agent` gives you fields to apply directly:
- `UserAgent` for `User-Agent` header.
- `Headers` and `HeaderOrder` for request header shaping.
- `ClientHelloID` / `ClientHelloSpec` for uTLS handshake selection.
- `H2Settings` for HTTP/2 transport tuning.

## Testing

```bash
go test ./...
```

## License

MIT. See `LICENSE`.
