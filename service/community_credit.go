package service

import (
	"math"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func ApplyChannelProviderPayout(relayInfo *relaycommon.RelayInfo, wallet *WalletFunding) error {
	if relayInfo == nil || wallet == nil || !common.CommunityCreditEnabled {
		return nil
	}
	if relayInfo.ChannelId <= 0 {
		return nil
	}
	dailyUsed := wallet.DailyConsumed()
	earnedUsed := wallet.EarnedConsumed()
	if dailyUsed <= 0 && earnedUsed <= 0 {
		return nil
	}

	channel, err := model.GetChannelById(relayInfo.ChannelId, false)
	if err != nil {
		return err
	}
	ownerUserId := channel.OwnerUserId
	if ownerUserId <= 0 {
		return nil
	}
	if ownerUserId == relayInfo.UserId && !common.ChannelPayoutSelfUseEnabled {
		return nil
	}

	dailyPayout := int(math.Round(float64(dailyUsed) * common.ChannelPayoutFromDailyRate))
	earnedPayout := int(math.Round(float64(earnedUsed) * common.ChannelPayoutFromEarnedRate))
	if dailyPayout <= 0 && earnedPayout <= 0 {
		return nil
	}

	_, dayStart, _ := model.CommunityDayWindow(relayInfo.StartTime)
	usage, err := model.GetChannelPayoutUsage(relayInfo.UserId, ownerUserId, relayInfo.ChannelId, dayStart)
	if err != nil {
		return err
	}

	consumerOwnerCap := model.CommunityDisplayAmountToQuotaUnits(common.ChannelPayoutConsumerOwnerDailyCap)
	consumerChannelCap := model.CommunityDisplayAmountToQuotaUnits(common.ChannelPayoutConsumerChannelDailyCap)
	ownerDailySourceCap := model.CommunityDisplayAmountToQuotaUnits(common.ChannelPayoutOwnerDailyDailySourceCap)

	totalPayout := dailyPayout + earnedPayout
	totalPayout = applyRemainingCap(totalPayout, consumerOwnerCap-usage.ConsumerOwnerAmount)
	totalPayout = applyRemainingCap(totalPayout, consumerChannelCap-usage.ConsumerChannelAmount)
	if totalPayout <= 0 {
		return nil
	}

	originalDailyPayout := dailyPayout
	originalTotal := dailyPayout + earnedPayout
	if originalTotal > totalPayout {
		dailyPayout = int(math.Round(float64(originalDailyPayout) * float64(totalPayout) / float64(originalTotal)))
		if dailyPayout > totalPayout {
			dailyPayout = totalPayout
		}
		earnedPayout = totalPayout - dailyPayout
	}

	if dailyPayout > 0 {
		dailyPayout = applyRemainingCap(dailyPayout, ownerDailySourceCap-usage.OwnerDailySourceAmount)
		if dailyPayout < 0 {
			dailyPayout = 0
		}
	}
	if dailyPayout+earnedPayout <= 0 {
		return nil
	}

	return model.CreateChannelPayoutCredit(ownerUserId, relayInfo.UserId, relayInfo.ChannelId, relayInfo.RequestId, dailyPayout, earnedPayout)
}

func ApplyTaskChannelProviderPayout(task *model.Task) error {
	if task == nil || task.PrivateData.BillingSource == BillingSourceSubscription {
		return nil
	}
	wallet := &WalletFunding{
		userId:         task.UserId,
		requestId:      task.PrivateData.RequestId,
		dailyConsumed:  task.PrivateData.DailyCreditConsumed,
		earnedConsumed: task.PrivateData.EarnedCreditConsumed,
		legacyConsumed: task.PrivateData.LegacyQuotaConsumed,
	}
	relayInfo := &relaycommon.RelayInfo{
		RequestId: task.PrivateData.RequestId,
		UserId:    task.UserId,
		StartTime: taskStartTime(task),
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelId: task.ChannelId,
		},
		ForcePreConsume: false,
	}
	return ApplyChannelProviderPayout(relayInfo, wallet)
}

func applyRemainingCap(amount int, remaining int) int {
	if amount <= 0 {
		return 0
	}
	if remaining <= 0 {
		return 0
	}
	if amount > remaining {
		return remaining
	}
	return amount
}

func taskStartTime(task *model.Task) time.Time {
	if task != nil && task.SubmitTime > 0 {
		return time.Unix(task.SubmitTime, 0)
	}
	return time.Now()
}
