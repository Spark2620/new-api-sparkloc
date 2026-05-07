package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"gorm.io/gorm"
)

type Checkin struct {
	Id           int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId       int    `json:"user_id" gorm:"not null;uniqueIndex:idx_user_checkin_date"`
	CheckinDate  string `json:"checkin_date" gorm:"type:varchar(10);not null;uniqueIndex:idx_user_checkin_date"`
	QuotaAwarded int    `json:"quota_awarded" gorm:"not null"`
	CreatedAt    int64  `json:"created_at" gorm:"bigint"`
}

type CheckinRecord struct {
	CheckinDate  string `json:"checkin_date"`
	QuotaAwarded int    `json:"quota_awarded"`
}

func (Checkin) TableName() string {
	return "checkins"
}

func GetUserCheckinRecords(userId int, startDate, endDate string) ([]Checkin, error) {
	var records []Checkin
	err := DB.Where("user_id = ? AND checkin_date >= ? AND checkin_date <= ?",
		userId, startDate, endDate).
		Order("checkin_date DESC").
		Find(&records).Error
	return records, err
}

func HasCheckedInToday(userId int) (bool, error) {
	today, _, _ := CommunityDayWindow(time.Now())
	var count int64
	err := DB.Model(&Checkin{}).
		Where("user_id = ? AND checkin_date = ?", userId, today).
		Count(&count).Error
	return count > 0, err
}

func UserCheckin(userId int) (*Checkin, error) {
	setting := operation_setting.GetCheckinSetting()
	if !setting.Enabled {
		return nil, errors.New("check-in is disabled")
	}

	hasChecked, err := HasCheckedInToday(userId)
	if err != nil {
		return nil, err
	}
	if hasChecked {
		return nil, errors.New("already checked in today")
	}

	user, err := GetUserById(userId, false)
	if err != nil {
		return nil, err
	}
	profile, err := GetCommunityProfileForUser(user)
	if err != nil {
		return nil, err
	}
	displayAmount, quotaAwarded := CalculateCommunityDailyReward(profile)
	if quotaAwarded <= 0 {
		return nil, errors.New("check-in reward is empty")
	}

	dayKey, _, expiresAt := CommunityDayWindow(time.Now())
	checkin := &Checkin{
		UserId:       userId,
		CheckinDate:  dayKey,
		QuotaAwarded: quotaAwarded,
		CreatedAt:    time.Now().Unix(),
	}
	meta := common.MapToJsonStr(map[string]interface{}{
		"display_amount":    displayAmount,
		"trust_level":       profile.TrustLevel,
		"leaderboard_score": profile.LeaderboardScore,
		"source":            profile.Source,
	})

	return userCheckinWithTransaction(checkin, userId, quotaAwarded, expiresAt, meta)
}

func userCheckinWithTransaction(checkin *Checkin, userId int, quotaAwarded int, expiresAt int64, meta string) (*Checkin, error) {
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(checkin).Error; err != nil {
			return errors.New("check-in failed, please try again later")
		}
		if err := CreateDailyCheckinCreditTx(tx, userId, checkin.CheckinDate, quotaAwarded, expiresAt, meta); err != nil {
			return errors.New("check-in failed: failed to grant credit")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return checkin, nil
}

func GetUserCheckinStats(userId int, month string) (map[string]interface{}, error) {
	startDate := month + "-01"
	endDate := month + "-31"

	records, err := GetUserCheckinRecords(userId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	checkinRecords := make([]CheckinRecord, len(records))
	for i, r := range records {
		checkinRecords[i] = CheckinRecord{
			CheckinDate:  r.CheckinDate,
			QuotaAwarded: r.QuotaAwarded,
		}
	}

	hasCheckedToday, _ := HasCheckedInToday(userId)

	var totalCheckins int64
	var totalQuota int64
	DB.Model(&Checkin{}).Where("user_id = ?", userId).Count(&totalCheckins)
	DB.Model(&Checkin{}).Where("user_id = ?", userId).Select("COALESCE(SUM(quota_awarded), 0)").Scan(&totalQuota)

	result := map[string]interface{}{
		"total_quota":      totalQuota,
		"total_checkins":   totalCheckins,
		"checkin_count":    len(records),
		"checked_in_today": hasCheckedToday,
		"records":          checkinRecords,
	}
	if balance, err := GetCommunityCreditBalance(userId); err == nil && balance != nil {
		result["daily_credit"] = balance.DailyCredit
		result["earned_credit"] = balance.EarnedCredit
		result["legacy_quota"] = balance.LegacyQuota
		result["available_quota"] = balance.TotalAvailableQuota
		result["daily_credit_expires_at"] = balance.DailyCreditExpiresAt
	}
	return result, nil
}
