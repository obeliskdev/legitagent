package legitagent

import (
	"math"

	"github.com/obeliskdev/fastrand"
	"golang.org/x/net/http2"
)

func randomizeValue(base uint32, percentage float64) uint32 {
	if base == 0 {
		return 0
	}

	delta := uint32(float64(base) * percentage)
	minX := uint32(1)
	if base > delta {
		minX = base - delta
	}

	maxX := base + delta
	return fastrand.Number(minX, maxX)
}

func clampRange(val, min, max uint32) uint32 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

type h2SettingRange struct {
	id      http2.SettingID
	percent float64
	min     uint32
	max     uint32
	clamp   bool
}

func (r h2SettingRange) apply(settings map[http2.SettingID]uint32, defaultVal uint32) {
	if val, ok := settings[r.id]; ok {
		val = randomizeValue(val, r.percent)
		if r.clamp {
			val = clampRange(val, r.min, r.max)
		}
		settings[r.id] = val
	} else if defaultVal != 0 {
		settings[r.id] = randomizeValue(defaultVal, r.percent)
	}
}

func randomizeH2Settings(baseSettings map[http2.SettingID]uint32, profile H2RandomizationProfile) map[http2.SettingID]uint32 {
	randomized := baseSettings

	switch profile {
	case H2RandomizationProfileNormal:
		h2SettingRange{id: http2.SettingHeaderTableSize, percent: 0.10}.apply(randomized, 0)
		h2SettingRange{id: http2.SettingInitialWindowSize, percent: 0.15, min: 65535, max: math.MaxInt32, clamp: true}.apply(randomized, 0)
		h2SettingRange{id: http2.SettingMaxFrameSize, percent: 0.05, min: 16384, max: 16777215, clamp: true}.apply(randomized, 0)
		h2SettingRange{id: http2.SettingMaxConcurrentStreams, percent: 0.10, min: 100, max: math.MaxUint32, clamp: true}.apply(randomized, 0)
		h2SettingRange{id: http2.SettingMaxHeaderListSize, percent: 0.10}.apply(randomized, 0)

	case H2RandomizationProfileMaximum:
		h2SettingRange{id: http2.SettingHeaderTableSize, percent: 0.30}.apply(randomized, 65536)
		randomized[http2.SettingEnablePush] = 0
		h2SettingRange{id: http2.SettingInitialWindowSize, percent: 0.35, min: 65535, max: math.MaxInt32, clamp: true}.apply(randomized, 6291456)
		h2SettingRange{id: http2.SettingMaxFrameSize, percent: 0.20, min: 16384, max: 16777215, clamp: true}.apply(randomized, 16384)
		h2SettingRange{id: http2.SettingMaxConcurrentStreams, percent: 0.50, min: 200, max: math.MaxUint32, clamp: true}.apply(randomized, 1000)
		h2SettingRange{id: http2.SettingMaxHeaderListSize, percent: 0.40}.apply(randomized, 0)
	default:
	}

	return randomized
}
