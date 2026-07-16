package legitagent

import (
	"net/http"
	"net/textproto"
)

var (
	hAccept                  = textproto.CanonicalMIMEHeaderKey("accept")
	hAcceptEncoding          = textproto.CanonicalMIMEHeaderKey("accept-encoding")
	hAcceptLanguage          = textproto.CanonicalMIMEHeaderKey("accept-language")
	hSecChUa                 = textproto.CanonicalMIMEHeaderKey("sec-ch-ua")
	hSecChUaArch             = textproto.CanonicalMIMEHeaderKey("sec-ch-ua-arch")
	hSecChUaBitness          = textproto.CanonicalMIMEHeaderKey("sec-ch-ua-bitness")
	hSecChUaFullVersionList  = textproto.CanonicalMIMEHeaderKey("sec-ch-ua-full-version-list")
	hSecChUaMobile           = textproto.CanonicalMIMEHeaderKey("sec-ch-ua-mobile")
	hSecChUaPlatform         = textproto.CanonicalMIMEHeaderKey("sec-ch-ua-platform")
	hSecChUaPlatformVersion  = textproto.CanonicalMIMEHeaderKey("sec-ch-ua-platform-version")
	hSecFetchDest            = textproto.CanonicalMIMEHeaderKey("sec-fetch-dest")
	hSecFetchMode            = textproto.CanonicalMIMEHeaderKey("sec-fetch-mode")
	hSecFetchSite            = textproto.CanonicalMIMEHeaderKey("sec-fetch-site")
	hSecFetchUser            = textproto.CanonicalMIMEHeaderKey("sec-fetch-user")
	hSecGpc                  = textproto.CanonicalMIMEHeaderKey("sec-gpc")
	hUpgradeInsecureRequests = textproto.CanonicalMIMEHeaderKey("upgrade-insecure-requests")
)

func headerSet(h http.Header, key, value string) {
	if existing, ok := h[key]; ok && cap(existing) > 0 {
		existing = existing[:1]
		existing[0] = value
		h[key] = existing
		return
	}
	h[key] = []string{value}
}
