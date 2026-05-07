package model

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	CreditTypeDaily  = "daily"
	CreditTypeEarned = "earned"

	CreditSourceDailyCheckin  = "daily_checkin"
	CreditSourceChannelPayout = "channel_payout"

	CreditTxnGrant         = "grant"
	CreditTxnConsume       = "consume"
	CreditTxnRefund        = "refund"
	CreditTxnChannelPayout = "channel_payout"
)

type CreditGrant struct {
	Id              int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId          int    `json:"user_id" gorm:"not null;index:idx_credit_grants_user_type_expire,priority:1;index"`
	CreditType      string `json:"credit_type" gorm:"type:varchar(20);not null;index:idx_credit_grants_user_type_expire,priority:2"`
	SourceType      string `json:"source_type" gorm:"type:varchar(40);not null;index"`
	Amount          int    `json:"amount" gorm:"not null"`
	RemainingAmount int    `json:"remaining_amount" gorm:"not null;index"`
	ExpiresAt       int64  `json:"expires_at" gorm:"bigint;default:0;index:idx_credit_grants_user_type_expire,priority:3"`
	DayKey          string `json:"day_key" gorm:"type:varchar(10);index"`
	SourceChannelId int    `json:"source_channel_id" gorm:"index"`
	SourceRequestId string `json:"source_request_id" gorm:"type:varchar(64);index"`
	SourceUserId    int    `json:"source_user_id" gorm:"index"`
	Meta            string `json:"meta" gorm:"type:text"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt       int64  `json:"updated_at" gorm:"bigint"`
}

func (CreditGrant) TableName() string {
	return "credit_grants"
}

type CreditConsumptionAllocation struct {
	Id             int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	RequestId      string `json:"request_id" gorm:"type:varchar(64);not null;index:idx_credit_alloc_request_user,priority:1;index"`
	UserId         int    `json:"user_id" gorm:"not null;index:idx_credit_alloc_request_user,priority:2;index"`
	GrantId        int64  `json:"grant_id" gorm:"not null;index"`
	CreditType     string `json:"credit_type" gorm:"type:varchar(20);not null"`
	Amount         int    `json:"amount" gorm:"not null"`
	RefundedAmount int    `json:"refunded_amount" gorm:"not null;default:0"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt      int64  `json:"updated_at" gorm:"bigint"`
}

func (CreditConsumptionAllocation) TableName() string {
	return "credit_consumption_allocations"
}

type CreditTransaction struct {
	Id                 int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId             int    `json:"user_id" gorm:"not null;index"`
	RequestId          string `json:"request_id" gorm:"type:varchar(64);index"`
	Type               string `json:"type" gorm:"type:varchar(32);not null;index"`
	SourceType         string `json:"source_type" gorm:"type:varchar(40);index"`
	Amount             int    `json:"amount" gorm:"not null"`
	DailyAmount        int    `json:"daily_amount" gorm:"not null;default:0"`
	EarnedAmount       int    `json:"earned_amount" gorm:"not null;default:0"`
	LegacyAmount       int    `json:"legacy_amount" gorm:"not null;default:0"`
	ChannelId          int    `json:"channel_id" gorm:"index"`
	ChannelOwnerUserId int    `json:"channel_owner_user_id" gorm:"index"`
	SourceUserId       int    `json:"source_user_id" gorm:"index"`
	DayKey             string `json:"day_key" gorm:"type:varchar(10);index"`
	CreatedAt          int64  `json:"created_at" gorm:"bigint;index"`
}

func (CreditTransaction) TableName() string {
	return "credit_transactions"
}

type CreditBalance struct {
	DailyCredit          int   `json:"daily_credit"`
	EarnedCredit         int   `json:"earned_credit"`
	LegacyQuota          int   `json:"legacy_quota"`
	TotalCommunityCredit int   `json:"total_community_credit"`
	TotalAvailableQuota  int   `json:"total_available_quota"`
	DailyCreditExpiresAt int64 `json:"daily_credit_expires_at"`
}

type CreditConsumeResult struct {
	DailyAmount  int
	EarnedAmount int
	TotalAmount  int
	Shortage     int
}

type ChannelPayoutUsage struct {
	ConsumerOwnerAmount    int
	ConsumerChannelAmount  int
	OwnerDailySourceAmount int
}

type CommunityProfile struct {
	TrustLevel       int
	LeaderboardScore int
	Source           string
}

func CommunityDisplayAmountToQuotaUnits(amount int) int {
	if amount <= 0 {
		return 0
	}
	return int(math.Round(float64(amount) * common.QuotaPerUnit))
}

func CommunityDayWindow(now time.Time) (dayKey string, startAt int64, endAt int64) {
	loc, err := time.LoadLocation(common.CommunityCreditTimezone)
	if err != nil {
		loc = time.Local
	}
	resetHour := common.CommunityCreditResetHour
	if resetHour < 0 || resetHour > 23 {
		resetHour = 4
	}
	localNow := now.In(loc)
	shifted := localNow.Add(-time.Duration(resetHour) * time.Hour)
	dayStart := time.Date(shifted.Year(), shifted.Month(), shifted.Day(), resetHour, 0, 0, 0, loc)
	dayEnd := dayStart.AddDate(0, 0, 1)
	return shifted.Format("2006-01-02"), dayStart.Unix(), dayEnd.Unix()
}

func CommunityCurrentDayKey() string {
	dayKey, _, _ := CommunityDayWindow(time.Now())
	return dayKey
}

func GetCommunityCreditBalance(userId int) (*CreditBalance, error) {
	balance := &CreditBalance{}
	legacyQuota, err := GetUserQuota(userId, false)
	if err != nil {
		return nil, err
	}
	balance.LegacyQuota = legacyQuota

	if common.CommunityCreditEnabled {
		now := time.Now().Unix()
		var rows []struct {
			CreditType string
			Total      int
		}
		if err := DB.Model(&CreditGrant{}).
			Select("credit_type, COALESCE(SUM(remaining_amount), 0) AS total").
			Where("user_id = ? AND remaining_amount > 0 AND (expires_at = 0 OR expires_at > ?)", userId, now).
			Group("credit_type").
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			switch row.CreditType {
			case CreditTypeDaily:
				balance.DailyCredit = row.Total
			case CreditTypeEarned:
				balance.EarnedCredit = row.Total
			}
		}
		var expiresAt int64
		_ = DB.Model(&CreditGrant{}).
			Select("COALESCE(MIN(expires_at), 0)").
			Where("user_id = ? AND credit_type = ? AND remaining_amount > 0 AND expires_at > ?", userId, CreditTypeDaily, now).
			Scan(&expiresAt).Error
		balance.DailyCreditExpiresAt = expiresAt
	}

	balance.TotalCommunityCredit = balance.DailyCredit + balance.EarnedCredit
	balance.TotalAvailableQuota = balance.TotalCommunityCredit + balance.LegacyQuota
	return balance, nil
}

func GetUserTotalWalletQuota(userId int) (int, error) {
	balance, err := GetCommunityCreditBalance(userId)
	if err != nil {
		return 0, err
	}
	return balance.TotalAvailableQuota, nil
}

func ConsumeCommunityCredits(requestId string, userId int, amount int) (*CreditConsumeResult, error) {
	result := &CreditConsumeResult{Shortage: amount}
	if amount <= 0 {
		result.Shortage = 0
		return result, nil
	}
	if !common.CommunityCreditEnabled {
		return result, nil
	}
	requestId = strings.TrimSpace(requestId)
	if requestId == "" {
		requestId = fmt.Sprintf("credit-%d-%d", userId, time.Now().UnixNano())
	}

	now := time.Now().Unix()
	err := DB.Transaction(func(tx *gorm.DB) error {
		remaining := amount
		for _, creditType := range []string{CreditTypeDaily, CreditTypeEarned} {
			if remaining <= 0 {
				break
			}
			var grants []CreditGrant
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("user_id = ? AND credit_type = ? AND remaining_amount > 0 AND (expires_at = 0 OR expires_at > ?)", userId, creditType, now).
				Order("expires_at ASC, id ASC").
				Find(&grants).Error; err != nil {
				return err
			}
			for _, grant := range grants {
				if remaining <= 0 {
					break
				}
				take := grant.RemainingAmount
				if take > remaining {
					take = remaining
				}
				if take <= 0 {
					continue
				}
				res := tx.Model(&CreditGrant{}).
					Where("id = ? AND remaining_amount >= ?", grant.Id, take).
					Updates(map[string]any{
						"remaining_amount": gorm.Expr("remaining_amount - ?", take),
						"updated_at":       time.Now().Unix(),
					})
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected != 1 {
					return errors.New("community credit changed concurrently")
				}
				allocation := CreditConsumptionAllocation{
					RequestId:  requestId,
					UserId:     userId,
					GrantId:    grant.Id,
					CreditType: creditType,
					Amount:     take,
					CreatedAt:  time.Now().Unix(),
					UpdatedAt:  time.Now().Unix(),
				}
				if err := tx.Create(&allocation).Error; err != nil {
					return err
				}
				switch creditType {
				case CreditTypeDaily:
					result.DailyAmount += take
				case CreditTypeEarned:
					result.EarnedAmount += take
				}
				result.TotalAmount += take
				remaining -= take
			}
		}
		result.Shortage = remaining
		if result.TotalAmount > 0 {
			dayKey, _, _ := CommunityDayWindow(time.Now())
			txn := CreditTransaction{
				UserId:       userId,
				RequestId:    requestId,
				Type:         CreditTxnConsume,
				SourceType:   "wallet",
				Amount:       result.TotalAmount,
				DailyAmount:  result.DailyAmount,
				EarnedAmount: result.EarnedAmount,
				DayKey:       dayKey,
				CreatedAt:    time.Now().Unix(),
			}
			if err := tx.Create(&txn).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func RefundCommunityCredits(requestId string, userId int, amount int) (*CreditConsumeResult, error) {
	result := &CreditConsumeResult{}
	if !common.CommunityCreditEnabled || amount == 0 {
		return result, nil
	}
	requestId = strings.TrimSpace(requestId)
	if requestId == "" {
		return nil, errors.New("request id is required")
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		remaining := amount
		var allocations []CreditConsumptionAllocation
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("request_id = ? AND user_id = ? AND amount > refunded_amount", requestId, userId).
			Order("id DESC")
		if err := query.Find(&allocations).Error; err != nil {
			return err
		}
		for _, allocation := range allocations {
			if remaining == 0 {
				break
			}
			available := allocation.Amount - allocation.RefundedAmount
			refund := available
			if remaining > 0 && refund > remaining {
				refund = remaining
			}
			if refund <= 0 {
				continue
			}
			res := tx.Model(&CreditConsumptionAllocation{}).
				Where("id = ? AND amount - refunded_amount >= ?", allocation.Id, refund).
				Updates(map[string]any{
					"refunded_amount": gorm.Expr("refunded_amount + ?", refund),
					"updated_at":      time.Now().Unix(),
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 1 {
				return errors.New("community credit allocation changed concurrently")
			}
			if err := tx.Model(&CreditGrant{}).
				Where("id = ?", allocation.GrantId).
				Updates(map[string]any{
					"remaining_amount": gorm.Expr("remaining_amount + ?", refund),
					"updated_at":       time.Now().Unix(),
				}).Error; err != nil {
				return err
			}
			switch allocation.CreditType {
			case CreditTypeDaily:
				result.DailyAmount += refund
			case CreditTypeEarned:
				result.EarnedAmount += refund
			}
			result.TotalAmount += refund
			if remaining > 0 {
				remaining -= refund
			}
		}
		if result.TotalAmount > 0 {
			dayKey, _, _ := CommunityDayWindow(time.Now())
			txn := CreditTransaction{
				UserId:       userId,
				RequestId:    requestId,
				Type:         CreditTxnRefund,
				SourceType:   "wallet",
				Amount:       result.TotalAmount,
				DailyAmount:  result.DailyAmount,
				EarnedAmount: result.EarnedAmount,
				DayKey:       dayKey,
				CreatedAt:    time.Now().Unix(),
			}
			if err := tx.Create(&txn).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func createCreditGrantTx(tx *gorm.DB, grant *CreditGrant) error {
	now := time.Now().Unix()
	if grant.Amount <= 0 {
		return nil
	}
	grant.RemainingAmount = grant.Amount
	if grant.CreatedAt == 0 {
		grant.CreatedAt = now
	}
	grant.UpdatedAt = now
	if err := tx.Create(grant).Error; err != nil {
		return err
	}
	txnType := CreditTxnGrant
	if grant.SourceType == CreditSourceChannelPayout {
		txnType = CreditTxnChannelPayout
	}
	txn := CreditTransaction{
		UserId:             grant.UserId,
		RequestId:          grant.SourceRequestId,
		Type:               txnType,
		SourceType:         grant.SourceType,
		Amount:             grant.Amount,
		DailyAmount:        0,
		EarnedAmount:       grant.Amount,
		ChannelId:          grant.SourceChannelId,
		ChannelOwnerUserId: grant.UserId,
		SourceUserId:       grant.SourceUserId,
		DayKey:             grant.DayKey,
		CreatedAt:          now,
	}
	if grant.CreditType == CreditTypeDaily {
		txn.DailyAmount = grant.Amount
		txn.EarnedAmount = 0
	}
	return tx.Create(&txn).Error
}

func CreateDailyCheckinCreditTx(tx *gorm.DB, userId int, dayKey string, amount int, expiresAt int64, meta string) error {
	return createCreditGrantTx(tx, &CreditGrant{
		UserId:          userId,
		CreditType:      CreditTypeDaily,
		SourceType:      CreditSourceDailyCheckin,
		Amount:          amount,
		ExpiresAt:       expiresAt,
		DayKey:          dayKey,
		SourceRequestId: fmt.Sprintf("checkin:%d:%s", userId, dayKey),
		Meta:            meta,
	})
}

func CreateChannelPayoutCredit(ownerUserId int, consumerUserId int, channelId int, requestId string, dailyPayout int, earnedPayout int) error {
	amount := dailyPayout + earnedPayout
	if !common.CommunityCreditEnabled || amount <= 0 {
		return nil
	}
	dayKey, _, _ := CommunityDayWindow(time.Now())
	var expiresAt int64
	if common.ChannelEarnedCreditTTLDays > 0 {
		expiresAt = time.Now().AddDate(0, 0, common.ChannelEarnedCreditTTLDays).Unix()
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var existing int64
		if err := tx.Model(&CreditTransaction{}).
			Where("type = ? AND source_type = ? AND request_id = ? AND user_id = ? AND channel_id = ?",
				CreditTxnChannelPayout, CreditSourceChannelPayout, requestId, ownerUserId, channelId).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return nil
		}
		grant := &CreditGrant{
			UserId:          ownerUserId,
			CreditType:      CreditTypeEarned,
			SourceType:      CreditSourceChannelPayout,
			Amount:          amount,
			ExpiresAt:       expiresAt,
			DayKey:          dayKey,
			SourceChannelId: channelId,
			SourceRequestId: requestId,
			SourceUserId:    consumerUserId,
		}
		now := time.Now().Unix()
		grant.RemainingAmount = amount
		grant.CreatedAt = now
		grant.UpdatedAt = now
		if err := tx.Create(grant).Error; err != nil {
			return err
		}
		txn := CreditTransaction{
			UserId:             ownerUserId,
			RequestId:          requestId,
			Type:               CreditTxnChannelPayout,
			SourceType:         CreditSourceChannelPayout,
			Amount:             amount,
			DailyAmount:        dailyPayout,
			EarnedAmount:       earnedPayout,
			ChannelId:          channelId,
			ChannelOwnerUserId: ownerUserId,
			SourceUserId:       consumerUserId,
			DayKey:             dayKey,
			CreatedAt:          now,
		}
		return tx.Create(&txn).Error
	})
}

func GetChannelPayoutUsage(consumerUserId int, ownerUserId int, channelId int, dayStart int64) (*ChannelPayoutUsage, error) {
	usage := &ChannelPayoutUsage{}
	if err := DB.Model(&CreditTransaction{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("type = ? AND source_type = ? AND source_user_id = ? AND channel_owner_user_id = ? AND created_at >= ?",
			CreditTxnChannelPayout, CreditSourceChannelPayout, consumerUserId, ownerUserId, dayStart).
		Scan(&usage.ConsumerOwnerAmount).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&CreditTransaction{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("type = ? AND source_type = ? AND source_user_id = ? AND channel_id = ? AND created_at >= ?",
			CreditTxnChannelPayout, CreditSourceChannelPayout, consumerUserId, channelId, dayStart).
		Scan(&usage.ConsumerChannelAmount).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&CreditTransaction{}).
		Select("COALESCE(SUM(daily_amount), 0)").
		Where("type = ? AND source_type = ? AND user_id = ? AND created_at >= ?",
			CreditTxnChannelPayout, CreditSourceChannelPayout, ownerUserId, dayStart).
		Scan(&usage.OwnerDailySourceAmount).Error; err != nil {
		return nil, err
	}
	return usage, nil
}

func GetCommunityProfileForUser(user *User) (*CommunityProfile, error) {
	profile := &CommunityProfile{TrustLevel: 0, LeaderboardScore: 0, Source: "fallback"}
	if user == nil {
		return profile, nil
	}
	profileURL := buildDiscourseURL(common.DiscourseProfileURLTemplate, user)
	if profileURL == "" && common.DiscourseBaseURL != "" && user.Username != "" {
		profileURL = common.DiscourseBaseURL + "/u/" + url.PathEscape(user.Username) + ".json"
	}
	if profileURL == "" {
		return profile, nil
	}
	body, err := fetchDiscourseJSON(profileURL)
	if err != nil {
		common.SysLog("failed to fetch discourse profile: " + err.Error())
		return profile, nil
	}
	profile.Source = "discourse"
	if trustLevel, ok := firstJSONInt(body, common.DiscourseTrustLevelJSONPath); ok {
		profile.TrustLevel = trustLevel
	}
	if score, ok := firstJSONInt(body, common.DiscourseLeaderboardScoreJSONPath); ok {
		profile.LeaderboardScore = score
	}
	leaderboardURL := buildDiscourseURL(common.DiscourseLeaderboardURLTemplate, user)
	if leaderboardURL != "" {
		leaderboardBody, err := fetchDiscourseJSON(leaderboardURL)
		if err != nil {
			common.SysLog("failed to fetch discourse leaderboard: " + err.Error())
		} else if score, ok := firstJSONInt(leaderboardBody, common.DiscourseLeaderboardScoreJSONPath); ok {
			profile.LeaderboardScore = score
		}
	}
	return profile, nil
}

func CalculateCommunityDailyReward(profile *CommunityProfile) (displayAmount int, quotaAmount int) {
	if profile == nil {
		profile = &CommunityProfile{}
	}
	tl := profile.TrustLevel
	if tl < 0 {
		tl = 0
	}
	tlBonus, ok := common.CommunityTLBonus[tl]
	if !ok {
		if tl > 4 {
			tlBonus = common.CommunityTLBonus[4]
		} else {
			tlBonus = common.CommunityTLBonus[0]
		}
	}
	leaderboardBonus := 0
	if common.CommunityLeaderboardStepPoints > 0 && profile.LeaderboardScore > 0 {
		leaderboardBonus = ((profile.LeaderboardScore - 1) / common.CommunityLeaderboardStepPoints) + 1
	}
	if common.CommunityLeaderboardBonusMax > 0 && leaderboardBonus > common.CommunityLeaderboardBonusMax {
		leaderboardBonus = common.CommunityLeaderboardBonusMax
	}
	displayAmount = tlBonus + leaderboardBonus
	if common.CommunityDailyCreditMax > 0 && displayAmount > common.CommunityDailyCreditMax {
		displayAmount = common.CommunityDailyCreditMax
	}
	quotaAmount = CommunityDisplayAmountToQuotaUnits(displayAmount)
	return displayAmount, quotaAmount
}

func buildDiscourseURL(template string, user *User) string {
	template = strings.TrimSpace(template)
	if template == "" || user == nil {
		return ""
	}
	replacements := map[string]string{
		"{user_id}":     strconv.Itoa(user.Id),
		"{username}":    user.Username,
		"{email}":       user.Email,
		"{sparkloc_id}": user.SparklocId,
	}
	result := template
	for key, value := range replacements {
		result = strings.ReplaceAll(result, key, url.QueryEscape(value))
	}
	return result
}

func fetchDiscourseJSON(target string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if common.DiscourseAPIKey != "" {
		req.Header.Set("Api-Key", common.DiscourseAPIKey)
	}
	if common.DiscourseAPIUsername != "" {
		req.Header.Set("Api-Username", common.DiscourseAPIUsername)
	}
	timeout := common.DiscourseTimeoutSeconds
	if timeout <= 0 {
		timeout = 10
	}
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("discourse returned status %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func firstJSONInt(body []byte, paths string) (int, bool) {
	for _, path := range strings.Split(paths, ",") {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		value := gjson.GetBytes(body, path)
		if !value.Exists() {
			continue
		}
		switch value.Type {
		case gjson.Number:
			return int(value.Int()), true
		case gjson.String:
			parsed, err := strconv.Atoi(strings.TrimSpace(value.String()))
			if err == nil {
				return parsed, true
			}
		case gjson.True:
			return 1, true
		case gjson.False:
			return 0, true
		}
	}
	return 0, false
}
