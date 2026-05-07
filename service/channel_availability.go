package service

import (
	"time"

	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func relayLatencyMs(info *relaycommon.RelayInfo) int64 {
	if info == nil {
		return 0
	}
	if info.HasSendResponse() {
		latency := info.FirstResponseTime.Sub(info.StartTime).Milliseconds()
		if latency >= 0 {
			return latency
		}
	}
	latency := time.Since(info.StartTime).Milliseconds()
	if latency < 0 {
		return 0
	}
	return latency
}

func RecordRelayChannelSuccess(info *relaycommon.RelayInfo) {
	if info == nil || info.ChannelMeta == nil || info.IsChannelTest {
		return
	}
	channelID := info.ChannelMeta.ChannelId
	if channelID <= 0 {
		return
	}
	if err := model.RecordChannelAvailabilitySuccess(channelID, relayLatencyMs(info)); err != nil {
		return
	}
}

func RecordRelayChannelFailure(info *relaycommon.RelayInfo) {
	if info == nil || info.ChannelMeta == nil || info.IsChannelTest {
		return
	}
	channelID := info.ChannelMeta.ChannelId
	if channelID <= 0 {
		return
	}
	if err := model.RecordChannelAvailabilityFailure(channelID, relayLatencyMs(info)); err != nil {
		return
	}
}
