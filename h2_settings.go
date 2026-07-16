package legitagent

import (
	"sync"

	"golang.org/x/net/http2"
)

var chromiumH2Settings = map[http2.SettingID]uint32{
	http2.SettingHeaderTableSize:      65536,
	http2.SettingEnablePush:           0,
	http2.SettingMaxConcurrentStreams: 1000,
	http2.SettingInitialWindowSize:    6291456,
	http2.SettingMaxFrameSize:         16384,
	http2.SettingMaxHeaderListSize:    262144,
}

var geckoH2Settings = map[http2.SettingID]uint32{
	http2.SettingHeaderTableSize:      65536,
	http2.SettingEnablePush:           0,
	http2.SettingMaxConcurrentStreams: 1000,
	http2.SettingInitialWindowSize:    131072,
	http2.SettingMaxFrameSize:         16384,
	http2.SettingMaxHeaderListSize:    262144,
}

var webkitH2Settings = map[http2.SettingID]uint32{
	http2.SettingHeaderTableSize:      4096,
	http2.SettingEnablePush:           0,
	http2.SettingMaxConcurrentStreams: 100,
	http2.SettingInitialWindowSize:    2097152,
	http2.SettingMaxFrameSize:         16384,
	http2.SettingMaxHeaderListSize:    16384,
}

var chromiumH2SettingSlice = h2SettingsToSlice(chromiumH2Settings)
var geckoH2SettingSlice = h2SettingsToSlice(geckoH2Settings)
var webkitH2SettingSlice = h2SettingsToSlice(webkitH2Settings)

func h2SettingsToSlice(m map[http2.SettingID]uint32) []http2.Setting {
	s := make([]http2.Setting, 0, len(m))
	for id, val := range m {
		s = append(s, http2.Setting{ID: id, Val: val})
	}
	return s
}

func ChromiumH2SettingSlice() []http2.Setting { return chromiumH2SettingSlice }
func GeckoH2SettingSlice() []http2.Setting    { return geckoH2SettingSlice }
func WebKitH2SettingSlice() []http2.Setting   { return webkitH2SettingSlice }

var (
	chromiumH2Pool = sync.Pool{
		New: func() any {
			m := make(map[http2.SettingID]uint32, len(chromiumH2Settings))
			return &m
		},
	}
	geckoH2Pool = sync.Pool{
		New: func() any {
			m := make(map[http2.SettingID]uint32, len(geckoH2Settings))
			return &m
		},
	}
	webkitH2Pool = sync.Pool{
		New: func() any {
			m := make(map[http2.SettingID]uint32, len(webkitH2Settings))
			return &m
		},
	}
)

func getH2Settings(pool *sync.Pool, template map[http2.SettingID]uint32) map[http2.SettingID]uint32 {
	mPtr := pool.Get().(*map[http2.SettingID]uint32)
	m := *mPtr
	for k := range m {
		delete(m, k)
	}
	for k, v := range template {
		m[k] = v
	}
	return m
}

func releaseH2Settings(pool *sync.Pool, m map[http2.SettingID]uint32) {
	for k := range m {
		delete(m, k)
	}
	pool.Put(&m)
}

func GetChromiumH2Settings() map[http2.SettingID]uint32 {
	return getH2Settings(&chromiumH2Pool, chromiumH2Settings)
}

func GetGeckoH2Settings() map[http2.SettingID]uint32 {
	return getH2Settings(&geckoH2Pool, geckoH2Settings)
}

func GetWebKitH2Settings() map[http2.SettingID]uint32 {
	return getH2Settings(&webkitH2Pool, webkitH2Settings)
}

func GetChromiumH2SettingsWithPool() (map[http2.SettingID]uint32, *sync.Pool) {
	return getH2Settings(&chromiumH2Pool, chromiumH2Settings), &chromiumH2Pool
}

func GetGeckoH2SettingsWithPool() (map[http2.SettingID]uint32, *sync.Pool) {
	return getH2Settings(&geckoH2Pool, geckoH2Settings), &geckoH2Pool
}

func GetWebKitH2SettingsWithPool() (map[http2.SettingID]uint32, *sync.Pool) {
	return getH2Settings(&webkitH2Pool, webkitH2Settings), &webkitH2Pool
}

func ReleaseChromiumH2Settings(m map[http2.SettingID]uint32) {
	releaseH2Settings(&chromiumH2Pool, m)
}

func ReleaseGeckoH2Settings(m map[http2.SettingID]uint32) {
	releaseH2Settings(&geckoH2Pool, m)
}

func ReleaseWebKitH2Settings(m map[http2.SettingID]uint32) {
	releaseH2Settings(&webkitH2Pool, m)
}
