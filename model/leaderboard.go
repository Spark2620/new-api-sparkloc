package model

import (
	"github.com/QuantumNous/new-api/common"
)

type LeaderboardType string

const (
	LeaderboardTypeConsumption  LeaderboardType = "consumption"
	LeaderboardTypeContribution LeaderboardType = "contribution"
)

type LeaderboardEntry struct {
	Rank          int    `json:"rank"`
	UserId        int    `json:"user_id"`
	Username      string `json:"username"`
	DisplayName   string `json:"display_name"`
	Role          int    `json:"role"`
	Group         string `json:"group"`
	Value         int64  `json:"value"`
	ValueLabel    string `json:"value_label"`
	IsCurrentUser bool   `json:"is_current_user"`
}

func GetLeaderboardEntries(leaderboardType LeaderboardType, currentUserId int, startIdx int, pageSize int) ([]*LeaderboardEntry, int64, error) {
	switch leaderboardType {
	case LeaderboardTypeContribution:
		return getContributionLeaderboardEntries(currentUserId, startIdx, pageSize)
	case LeaderboardTypeConsumption:
		fallthrough
	default:
		return getConsumptionLeaderboardEntries(currentUserId, startIdx, pageSize)
	}
}

func getConsumptionLeaderboardEntries(currentUserId int, startIdx int, pageSize int) ([]*LeaderboardEntry, int64, error) {
	baseQuery := DB.Model(&User{}).
		Where("status = ? AND used_quota > 0", common.UserStatusEnabled)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]*LeaderboardEntry, 0)
	err := DB.Model(&User{}).
		Select("id AS user_id, username, display_name, role, `group`, used_quota AS value").
		Where("status = ? AND used_quota > 0", common.UserStatusEnabled).
		Order("used_quota DESC, id ASC").
		Offset(startIdx).
		Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	for i, row := range rows {
		row.Rank = startIdx + i + 1
		row.ValueLabel = "used_quota"
		row.IsCurrentUser = row.UserId == currentUserId
	}

	return rows, total, nil
}

func getContributionLeaderboardEntries(currentUserId int, startIdx int, pageSize int) ([]*LeaderboardEntry, int64, error) {
	type contributionRow struct {
		UserId      int    `json:"user_id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Role        int    `json:"role"`
		Group       string `json:"group"`
		Value       int64  `json:"value"`
	}

	baseQuery := DB.Table("credit_transactions AS ct").
		Joins("JOIN users AS u ON u.id = ct.user_id").
		Where("ct.type = ? AND ct.source_type = ? AND u.status = ? AND u.deleted_at IS NULL", CreditTxnChannelPayout, CreditSourceChannelPayout, common.UserStatusEnabled).
		Group("u.id, u.username, u.display_name, u.role, u.`group`")

	var total int64
	if err := DB.Table("(?) AS contribution_rows", baseQuery.Select("u.id AS user_id")).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rawRows []*contributionRow
	err := baseQuery.
		Select("u.id AS user_id, u.username, u.display_name, u.role, u.`group`, COALESCE(SUM(ct.amount), 0) AS value").
		Order("value DESC, u.id ASC").
		Offset(startIdx).
		Limit(pageSize).
		Scan(&rawRows).Error
	if err != nil {
		return nil, 0, err
	}

	rows := make([]*LeaderboardEntry, 0, len(rawRows))
	for i, row := range rawRows {
		rows = append(rows, &LeaderboardEntry{
			Rank:          startIdx + i + 1,
			UserId:        row.UserId,
			Username:      row.Username,
			DisplayName:   row.DisplayName,
			Role:          row.Role,
			Group:         row.Group,
			Value:         row.Value,
			ValueLabel:    "channel_payout",
			IsCurrentUser: row.UserId == currentUserId,
		})
	}

	return rows, total, nil
}
