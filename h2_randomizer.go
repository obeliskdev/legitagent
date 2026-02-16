package legitagent

import (
	"github.com/obeliskdev/fastrand"
	"golang.org/x/net/http2"
	"math"
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

func randomizeH2Settings(baseSettings map[http2.SettingID]uint32, profile H2RandomizationProfile) map[http2.SettingID]uint32 {
	randomized := make(map[http2.SettingID]uint32)
	for id, val := range baseSettings {
		randomized[id] = val
	}

	switch profile {
	case H2RandomizationProfileNormal:
		if val, ok := randomized[http2.SettingHeaderTableSize]; ok {
			randomized[http2.SettingHeaderTableSize] = randomizeValue(val, 0.10)
		}
		if val, ok := randomized[http2.SettingInitialWindowSize]; ok {
			randomized[http2.SettingInitialWindowSize] = clampRange(randomizeValue(val, 0.15), 65535, math.MaxInt32)
		}
		if val, ok := randomized[http2.SettingMaxFrameSize]; ok {
			randomized[http2.SettingMaxFrameSize] = clampRange(randomizeValue(val, 0.05), 16384, 16777215)
		}
		if val, ok := randomized[http2.SettingMaxConcurrentStreams]; ok {
			randomized[http2.SettingMaxConcurrentStreams] = clampRange(randomizeValue(val, 0.10), 100, math.MaxUint32)
		}
		if val, ok := randomized[http2.SettingMaxHeaderListSize]; ok {
			randomized[http2.SettingMaxHeaderListSize] = randomizeValue(val, 0.10)
		}

	case H2RandomizationProfileMaximum:
		if val, ok := randomized[http2.SettingHeaderTableSize]; ok {
			randomized[http2.SettingHeaderTableSize] = randomizeValue(val, 0.30)
		} else {
			randomized[http2.SettingHeaderTableSize] = randomizeValue(65536, 0.30)
		}
		randomized[http2.SettingEnablePush] = 0

		if val, ok := randomized[http2.SettingInitialWindowSize]; ok {
			randomized[http2.SettingInitialWindowSize] = clampRange(randomizeValue(val, 0.35), 65535, math.MaxInt32)
		} else {
			randomized[http2.SettingInitialWindowSize] = randomizeValue(6291456, 0.35)
		}

		if val, ok := randomized[http2.SettingMaxFrameSize]; ok {
			randomized[http2.SettingMaxFrameSize] = clampRange(randomizeValue(val, 0.20), 16384, 16777215)
		} else {
			randomized[http2.SettingMaxFrameSize] = randomizeValue(16384, 0.20)
		}

		if val, ok := randomized[http2.SettingMaxConcurrentStreams]; ok {
			randomized[http2.SettingMaxConcurrentStreams] = clampRange(randomizeValue(val, 0.50), 200, math.MaxUint32)
		} else {
			randomized[http2.SettingMaxConcurrentStreams] = randomizeValue(1000, 0.50)
		}

		if val, ok := randomized[http2.SettingMaxHeaderListSize]; ok {
			randomized[http2.SettingMaxHeaderListSize] = randomizeValue(val, 0.40)
		}
	default:
	}

	return randomized
}
