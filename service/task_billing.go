package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
)

func LogTaskConsumption(c *gin.Context, info *relaycommon.RelayInfo) {
	tokenName := c.GetString("token_name")
	logContent := fmt.Sprintf("action %s", info.Action)
	if common.StringsContains(constant.TaskPricePatches, info.OriginModelName) {
		logContent = fmt.Sprintf("%s, per-call billing", logContent)
	} else if len(info.PriceData.OtherRatios) > 0 {
		var contents []string
		for key, ratio := range info.PriceData.OtherRatios {
			if ratio != 1.0 {
				contents = append(contents, fmt.Sprintf("%s: %.2f", key, ratio))
			}
		}
		if len(contents) > 0 {
			logContent = fmt.Sprintf("%s, ratios: %s", logContent, strings.Join(contents, ", "))
		}
	}

	other := map[string]interface{}{
		"is_task":      true,
		"request_path": c.Request.URL.Path,
		"model_price":  info.PriceData.ModelPrice,
		"group_ratio":  info.PriceData.GroupRatioInfo.GroupRatio,
	}
	if info.PriceData.ModelRatio > 0 {
		other["model_ratio"] = info.PriceData.ModelRatio
	}
	if info.PriceData.GroupRatioInfo.HasSpecialRatio {
		other["user_group_ratio"] = info.PriceData.GroupRatioInfo.GroupSpecialRatio
	}
	if info.IsModelMapped {
		other["is_model_mapped"] = true
		other["upstream_model_name"] = info.UpstreamModelName
	}

	model.RecordConsumeLog(c, info.UserId, model.RecordConsumeLogParams{
		ChannelId: info.ChannelId,
		ModelName: info.OriginModelName,
		TokenName: tokenName,
		Quota:     info.PriceData.Quota,
		Content:   logContent,
		TokenId:   info.TokenId,
		Group:     info.UsingGroup,
		Other:     other,
	})
	model.UpdateUserUsedQuotaAndRequestCount(info.UserId, info.PriceData.Quota)
	model.UpdateChannelUsedQuota(info.ChannelId, info.PriceData.Quota)
}

func resolveTokenKey(ctx context.Context, tokenId int, taskID string) string {
	token, err := model.GetTokenById(tokenId)
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("failed to resolve token key (tokenId=%d, task=%s): %s", tokenId, taskID, err.Error()))
		return ""
	}
	return token.Key
}

func taskIsSubscription(task *model.Task) bool {
	return task.PrivateData.BillingSource == BillingSourceSubscription && task.PrivateData.SubscriptionId > 0
}

func ensureTaskRequestID(task *model.Task) string {
	if task == nil {
		return ""
	}
	requestID := strings.TrimSpace(task.PrivateData.RequestId)
	if requestID != "" {
		return requestID
	}
	requestID = fmt.Sprintf("task:%s", task.TaskID)
	task.PrivateData.RequestId = requestID
	return requestID
}

func adjustTaskWalletFunding(task *model.Task, delta int) error {
	if task == nil || delta == 0 {
		return nil
	}

	requestID := ensureTaskRequestID(task)
	if delta > 0 {
		community, err := model.ConsumeCommunityCredits(requestID, task.UserId, delta)
		if err != nil {
			return err
		}

		legacyAmount := delta
		if community != nil {
			task.PrivateData.DailyCreditConsumed += community.DailyAmount
			task.PrivateData.EarnedCreditConsumed += community.EarnedAmount
			legacyAmount = community.Shortage
		}

		if legacyAmount > 0 {
			if err := model.DecreaseUserQuota(task.UserId, legacyAmount, false); err != nil {
				if community != nil && community.TotalAmount > 0 {
					_, _ = model.RefundCommunityCredits(requestID, task.UserId, community.TotalAmount)
					task.PrivateData.DailyCreditConsumed -= community.DailyAmount
					if task.PrivateData.DailyCreditConsumed < 0 {
						task.PrivateData.DailyCreditConsumed = 0
					}
					task.PrivateData.EarnedCreditConsumed -= community.EarnedAmount
					if task.PrivateData.EarnedCreditConsumed < 0 {
						task.PrivateData.EarnedCreditConsumed = 0
					}
				}
				return err
			}
			task.PrivateData.LegacyQuotaConsumed += legacyAmount
		}
		return nil
	}

	remaining := -delta
	refundLegacy := task.PrivateData.LegacyQuotaConsumed
	if refundLegacy > remaining {
		refundLegacy = remaining
	}
	if refundLegacy > 0 {
		if err := model.IncreaseUserQuota(task.UserId, refundLegacy, false); err != nil {
			return err
		}
		task.PrivateData.LegacyQuotaConsumed -= refundLegacy
		remaining -= refundLegacy
	}

	if remaining <= 0 {
		return nil
	}

	community, err := model.RefundCommunityCredits(requestID, task.UserId, remaining)
	if err != nil {
		return err
	}
	if community != nil {
		task.PrivateData.DailyCreditConsumed -= community.DailyAmount
		if task.PrivateData.DailyCreditConsumed < 0 {
			task.PrivateData.DailyCreditConsumed = 0
		}
		task.PrivateData.EarnedCreditConsumed -= community.EarnedAmount
		if task.PrivateData.EarnedCreditConsumed < 0 {
			task.PrivateData.EarnedCreditConsumed = 0
		}
	}
	return nil
}

func taskAdjustFunding(task *model.Task, delta int) error {
	if task == nil || delta == 0 {
		return nil
	}
	if taskIsSubscription(task) {
		return model.PostConsumeUserSubscriptionDelta(task.PrivateData.SubscriptionId, int64(delta))
	}
	return adjustTaskWalletFunding(task, delta)
}

func taskAdjustTokenQuota(ctx context.Context, task *model.Task, delta int) {
	if task == nil || task.PrivateData.TokenId <= 0 || delta == 0 {
		return
	}

	tokenKey := resolveTokenKey(ctx, task.PrivateData.TokenId, task.TaskID)
	if tokenKey == "" {
		return
	}

	var err error
	if delta > 0 {
		err = model.DecreaseTokenQuota(task.PrivateData.TokenId, tokenKey, delta)
	} else {
		err = model.IncreaseTokenQuota(task.PrivateData.TokenId, tokenKey, -delta)
	}
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("failed to adjust token quota (delta=%d, task=%s): %s", delta, task.TaskID, err.Error()))
	}
}

func persistTaskBillingState(ctx context.Context, task *model.Task) {
	if task == nil {
		return
	}
	if err := task.Update(); err != nil {
		logger.LogError(ctx, fmt.Sprintf("failed to persist task billing state (task=%s): %s", task.TaskID, err.Error()))
	}
}

func taskBillingOther(task *model.Task) map[string]interface{} {
	other := make(map[string]interface{})
	if bc := task.PrivateData.BillingContext; bc != nil {
		other["model_price"] = bc.ModelPrice
		if bc.ModelRatio > 0 {
			other["model_ratio"] = bc.ModelRatio
		}
		other["group_ratio"] = bc.GroupRatio
		for k, v := range bc.OtherRatios {
			other[k] = v
		}
	}

	props := task.Properties
	if props.UpstreamModelName != "" && props.UpstreamModelName != props.OriginModelName {
		other["is_model_mapped"] = true
		other["upstream_model_name"] = props.UpstreamModelName
	}
	return other
}

func taskModelName(task *model.Task) string {
	if bc := task.PrivateData.BillingContext; bc != nil && bc.OriginModelName != "" {
		return bc.OriginModelName
	}
	return task.Properties.OriginModelName
}

func RefundTaskQuota(ctx context.Context, task *model.Task, reason string) {
	if task == nil || task.Quota == 0 {
		return
	}

	quota := task.Quota
	if err := taskAdjustFunding(task, -quota); err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("failed to refund task funding (task=%s): %s", task.TaskID, err.Error()))
		return
	}
	taskAdjustTokenQuota(ctx, task, -quota)
	persistTaskBillingState(ctx, task)

	other := taskBillingOther(task)
	other["task_id"] = task.TaskID
	other["reason"] = reason
	model.RecordTaskBillingLog(model.RecordTaskBillingLogParams{
		UserId:    task.UserId,
		LogType:   model.LogTypeRefund,
		Content:   "",
		ChannelId: task.ChannelId,
		ModelName: taskModelName(task),
		Quota:     quota,
		TokenId:   task.PrivateData.TokenId,
		Group:     task.Group,
		Other:     other,
	})
}

func RecalculateTaskQuota(ctx context.Context, task *model.Task, actualQuota int, reason string) {
	if task == nil || actualQuota <= 0 {
		return
	}

	preConsumedQuota := task.Quota
	quotaDelta := actualQuota - preConsumedQuota
	if quotaDelta == 0 {
		logger.LogInfo(ctx, fmt.Sprintf("task %s quota unchanged (%s): %s", task.TaskID, logger.LogQuota(actualQuota), reason))
		return
	}

	logger.LogInfo(ctx, fmt.Sprintf(
		"task %s billing delta=%s (actual=%s, pre=%s): %s",
		task.TaskID,
		logger.LogQuota(quotaDelta),
		logger.LogQuota(actualQuota),
		logger.LogQuota(preConsumedQuota),
		reason,
	))

	if err := taskAdjustFunding(task, quotaDelta); err != nil {
		logger.LogError(ctx, fmt.Sprintf("failed to settle task funding delta (task=%s): %s", task.TaskID, err.Error()))
		return
	}
	taskAdjustTokenQuota(ctx, task, quotaDelta)

	task.Quota = actualQuota
	persistTaskBillingState(ctx, task)

	var logType int
	var logQuota int
	if quotaDelta > 0 {
		logType = model.LogTypeConsume
		logQuota = quotaDelta
		model.UpdateUserUsedQuotaAndRequestCount(task.UserId, quotaDelta)
		model.UpdateChannelUsedQuota(task.ChannelId, quotaDelta)
	} else {
		logType = model.LogTypeRefund
		logQuota = -quotaDelta
	}

	other := taskBillingOther(task)
	other["task_id"] = task.TaskID
	other["pre_consumed_quota"] = preConsumedQuota
	other["actual_quota"] = actualQuota
	model.RecordTaskBillingLog(model.RecordTaskBillingLogParams{
		UserId:    task.UserId,
		LogType:   logType,
		Content:   reason,
		ChannelId: task.ChannelId,
		ModelName: taskModelName(task),
		Quota:     logQuota,
		TokenId:   task.PrivateData.TokenId,
		Group:     task.Group,
		Other:     other,
	})
}

func RecalculateTaskQuotaByTokens(ctx context.Context, task *model.Task, totalTokens int) {
	if task == nil || totalTokens <= 0 {
		return
	}

	modelName := taskModelName(task)
	modelRatio, hasRatioSetting, _ := ratio_setting.GetModelRatio(modelName)
	if !hasRatioSetting || modelRatio <= 0 {
		return
	}

	group := task.Group
	if group == "" {
		user, err := model.GetUserById(task.UserId, false)
		if err == nil {
			group = user.Group
		}
	}
	if group == "" {
		return
	}

	groupRatio := ratio_setting.GetGroupRatio(group)
	userGroupRatio, hasUserGroupRatio := ratio_setting.GetGroupGroupRatio(group, group)
	finalGroupRatio := groupRatio
	if hasUserGroupRatio {
		finalGroupRatio = userGroupRatio
	}

	otherMultiplier := 1.0
	if bc := task.PrivateData.BillingContext; bc != nil {
		for _, ratio := range bc.OtherRatios {
			if ratio > 0 && ratio != 1.0 {
				otherMultiplier *= ratio
			}
		}
	}

	actualQuota := int(float64(totalTokens) * modelRatio * finalGroupRatio * otherMultiplier)
	reason := fmt.Sprintf(
		"token recalculation: tokens=%d, modelRatio=%.2f, groupRatio=%.2f, otherMultiplier=%.4f",
		totalTokens,
		modelRatio,
		finalGroupRatio,
		otherMultiplier,
	)
	RecalculateTaskQuota(ctx, task, actualQuota, reason)
}
