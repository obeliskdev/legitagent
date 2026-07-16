package legitagent

import "golang.org/x/net/http2"

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

func GetChromiumH2Settings() map[http2.SettingID]uint32 {
	out := make(map[http2.SettingID]uint32, len(chromiumH2Settings))
	for k, v := range chromiumH2Settings {
		out[k] = v
	}
	return out
}

func GetGeckoH2Settings() map[http2.SettingID]uint32 {
	out := make(map[http2.SettingID]uint32, len(geckoH2Settings))
	for k, v := range geckoH2Settings {
		out[k] = v
	}
	return out
}

func GetWebKitH2Settings() map[http2.SettingID]uint32 {
	out := make(map[http2.SettingID]uint32, len(webkitH2Settings))
	for k, v := range webkitH2Settings {
		out[k] = v
	}
	return out
}
