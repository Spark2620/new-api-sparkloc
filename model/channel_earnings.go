package model

import (
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"gorm.io/gorm"
)

type ChannelEarningsItem struct {
	CreatedAt         int64  `json:"created_at"`
	ChannelID         int    `json:"channel_id"`
	ChannelName       string `json:"channel_name"`
	Consumer          string `json:"consumer"`
	SelfUse           bool   `json:"self_use"`
	ModelName         string `json:"model_name"`
	Quota             int    `json:"quota"`
	PromptTokens      int    `json:"prompt_tokens"`
	CompletionTokens  int    `json:"completion_tokens"`
	TotalTokens       int    `json:"total_tokens"`
	PayoutAmount      int64  `json:"payout_amount"`
	PayoutDailyAmount int64  `json:"payout_daily_amount"`
	PayoutEarnedAmount int64 `json:"payout_earned_amount"`
}

type ChannelEarningsSummary struct {
	OwnedChannels      int64 `json:"owned_channels"`
	MatchedRequests    int64 `json:"matched_requests"`
	MatchedConsumption int64 `json:"matched_consumption"`
	TotalEarnings      int64 `json:"total_earnings"`
}

type ownedChannelSummaryRow struct {
	ID   int    `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

type channelPayoutAggregateRow struct {
	RequestID    string `gorm:"column:request_id"`
	ChannelID    int    `gorm:"column:channel_id"`
	SourceUserID int    `gorm:"column:source_user_id"`
	Amount       int64  `gorm:"column:amount"`
	DailyAmount  int64  `gorm:"column:daily_amount"`
	EarnedAmount int64  `gorm:"column:earned_amount"`
}

type channelPayoutAggregate struct {
	Amount       int64
	DailyAmount  int64
	EarnedAmount int64
}

func GetChannelEarnings(ownerUserID int, keyword string, startIdx int, pageSize int) ([]*ChannelEarningsItem, *ChannelEarningsSummary, error) {
	if pageSize <= 0 {
		pageSize = common.ItemsPerPage
	}

	summary := &ChannelEarningsSummary{}
	ownedChannels, err := getOwnedChannelsForEarnings(ownerUserID)
	if err != nil {
		return nil, nil, err
	}
	summary.OwnedChannels = int64(len(ownedChannels))

	if err := DB.Model(&CreditTransaction{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("channel_owner_user_id = ? AND type = ? AND source_type = ?",
			ownerUserID,
			CreditTxnChannelPayout,
			CreditSourceChannelPayout,
		).
		Scan(&summary.TotalEarnings).Error; err != nil {
		return nil, nil, err
	}

	if len(ownedChannels) == 0 {
		return []*ChannelEarningsItem{}, summary, nil
	}

	channelIDs := make([]int, 0, len(ownedChannels))
	channelNameMap := make(map[int]string, len(ownedChannels))
	for _, channel := range ownedChannels {
		channelIDs = append(channelIDs, channel.ID)
		channelNameMap[channel.ID] = channel.Name
	}

	filteredQuery := buildChannelEarningsLogsQuery(channelIDs, ownedChannels, keyword)

	if err := filteredQuery.Count(&summary.MatchedRequests).Error; err != nil {
		return nil, nil, err
	}
	if err := buildChannelEarningsLogsQuery(channelIDs, ownedChannels, keyword).
		Select("COALESCE(SUM(logs.quota), 0)").
		Scan(&summary.MatchedConsumption).Error; err != nil {
		return nil, nil, err
	}
	if summary.MatchedRequests == 0 {
		return []*ChannelEarningsItem{}, summary, nil
	}

	var logs []*Log
	if err := buildChannelEarningsLogsQuery(channelIDs, ownedChannels, keyword).
		Order("logs.created_at DESC, logs.id DESC").
		Offset(startIdx).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, nil, err
	}

	payoutMap, err := getChannelPayoutAggregates(ownerUserID, logs)
	if err != nil {
		return nil, nil, err
	}

	items := make([]*ChannelEarningsItem, 0, len(logs))
	for _, log := range logs {
		if log == nil {
			continue
		}
		key := buildChannelPayoutAggregateKey(log.RequestId, log.ChannelId, log.UserId)
		payout := payoutMap[key]
		items = append(items, &ChannelEarningsItem{
			CreatedAt:          log.CreatedAt,
			ChannelID:          log.ChannelId,
			ChannelName:        channelNameMap[log.ChannelId],
			Consumer:           maskChannelConsumer(log.Username),
			SelfUse:            log.UserId == ownerUserID,
			ModelName:          log.ModelName,
			Quota:              log.Quota,
			PromptTokens:       log.PromptTokens,
			CompletionTokens:   log.CompletionTokens,
			TotalTokens:        log.PromptTokens + log.CompletionTokens,
			PayoutAmount:       payout.Amount,
			PayoutDailyAmount:  payout.DailyAmount,
			PayoutEarnedAmount: payout.EarnedAmount,
		})
	}

	return items, summary, nil
}

func getOwnedChannelsForEarnings(ownerUserID int) ([]ownedChannelSummaryRow, error) {
	rows := make([]ownedChannelSummaryRow, 0)
	err := DB.Model(&Channel{}).
		Select("id, name").
		Where("owner_user_id = ?", ownerUserID).
		Where("type IN ?", constant.SupportedChannelTypeIDs).
		Order("id DESC").
		Find(&rows).Error
	return rows, err
}

func buildChannelEarningsLogsQuery(channelIDs []int, ownedChannels []ownedChannelSummaryRow, keyword string) *gorm.DB {
	tx := LOG_DB.Model(&Log{}).
		Where("logs.type = ?", LogTypeConsume).
		Where("logs.channel_id IN ?", channelIDs)

	trimmedKeyword := strings.TrimSpace(keyword)
	if trimmedKeyword == "" {
		return tx
	}

	likeKeyword := "%" + escapeLikeKeyword(trimmedKeyword) + "%"
	conditions := []string{
		"logs.username LIKE ? ESCAPE '!'",
		"logs.model_name LIKE ? ESCAPE '!'",
		"logs.request_id = ?",
	}
	args := []any{likeKeyword, likeKeyword, trimmedKeyword}

	matchedChannelIDs := filterOwnedChannelIDsByKeyword(ownedChannels, trimmedKeyword)
	if len(matchedChannelIDs) > 0 {
		conditions = append(conditions, "logs.channel_id IN ?")
		args = append(args, matchedChannelIDs)
	}

	return tx.Where("("+strings.Join(conditions, " OR ")+")", args...)
}

func filterOwnedChannelIDsByKeyword(ownedChannels []ownedChannelSummaryRow, keyword string) []int {
	if len(ownedChannels) == 0 {
		return nil
	}
	lowerKeyword := strings.ToLower(strings.TrimSpace(keyword))
	if lowerKeyword == "" {
		return nil
	}

	channelID, hasChannelID := parseChannelSearchID(lowerKeyword)
	matchedIDs := make([]int, 0)
	for _, channel := range ownedChannels {
		if hasChannelID && channel.ID == channelID {
			matchedIDs = append(matchedIDs, channel.ID)
			continue
		}
		if strings.Contains(strings.ToLower(channel.Name), lowerKeyword) {
			matchedIDs = append(matchedIDs, channel.ID)
		}
	}
	return matchedIDs
}

func getChannelPayoutAggregates(ownerUserID int, logs []*Log) (map[string]channelPayoutAggregate, error) {
	requestIDs := make([]string, 0, len(logs))
	channelIDs := make([]int, 0, len(logs))
	seenRequestID := make(map[string]struct{}, len(logs))
	seenChannelID := make(map[int]struct{}, len(logs))

	for _, log := range logs {
		if log == nil || strings.TrimSpace(log.RequestId) == "" {
			continue
		}
		if _, ok := seenRequestID[log.RequestId]; !ok {
			seenRequestID[log.RequestId] = struct{}{}
			requestIDs = append(requestIDs, log.RequestId)
		}
		if _, ok := seenChannelID[log.ChannelId]; !ok {
			seenChannelID[log.ChannelId] = struct{}{}
			channelIDs = append(channelIDs, log.ChannelId)
		}
	}
	if len(requestIDs) == 0 || len(channelIDs) == 0 {
		return map[string]channelPayoutAggregate{}, nil
	}

	rows := make([]channelPayoutAggregateRow, 0)
	if err := DB.Model(&CreditTransaction{}).
		Select(`
			request_id,
			channel_id,
			source_user_id,
			COALESCE(SUM(amount), 0) AS amount,
			COALESCE(SUM(daily_amount), 0) AS daily_amount,
			COALESCE(SUM(earned_amount), 0) AS earned_amount
		`).
		Where("channel_owner_user_id = ? AND type = ? AND source_type = ?", ownerUserID, CreditTxnChannelPayout, CreditSourceChannelPayout).
		Where("request_id IN ?", requestIDs).
		Where("channel_id IN ?", channelIDs).
		Group("request_id, channel_id, source_user_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[string]channelPayoutAggregate, len(rows))
	for _, row := range rows {
		result[buildChannelPayoutAggregateKey(row.RequestID, row.ChannelID, row.SourceUserID)] = channelPayoutAggregate{
			Amount:       row.Amount,
			DailyAmount:  row.DailyAmount,
			EarnedAmount: row.EarnedAmount,
		}
	}
	return result, nil
}

func buildChannelPayoutAggregateKey(requestID string, channelID int, sourceUserID int) string {
	return requestID + ":" + strconv.Itoa(channelID) + ":" + strconv.Itoa(sourceUserID)
}

func maskChannelConsumer(username string) string {
	trimmed := strings.TrimSpace(username)
	if trimmed == "" {
		return "****"
	}
	runes := []rune(trimmed)
	switch len(runes) {
	case 1:
		return string(runes[0]) + "*"
	case 2:
		return string(runes[0]) + "*"
	default:
		return string(runes[0]) + "***" + string(runes[len(runes)-1])
	}
}
