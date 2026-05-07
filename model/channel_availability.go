package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
)

const channelAvailabilityWindowMinutes = 60

type ChannelAvailabilityBucket struct {
	ID             int64 `json:"id"`
	ChannelID      int   `json:"channel_id" gorm:"index:idx_channel_bucket,unique"`
	BucketStart    int64 `json:"bucket_start" gorm:"bigint;index:idx_channel_bucket,unique;index"`
	SuccessCount   int   `json:"success_count" gorm:"default:0"`
	FailureCount   int   `json:"failure_count" gorm:"default:0"`
	LatencyTotalMs int64 `json:"latency_total_ms" gorm:"bigint;default:0"`
	UpdatedAt      int64 `json:"updated_at" gorm:"bigint"`
}

type ChannelAvailabilityMinute struct {
	BucketStart   int64   `json:"bucket_start"`
	SuccessCount  int     `json:"success_count"`
	FailureCount  int     `json:"failure_count"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	Availability  float64 `json:"availability"`
	Status        string  `json:"status"`
	HasTraffic    bool    `json:"has_traffic"`
	TotalRequests int     `json:"total_requests"`
}

type ChannelAvailabilityItem struct {
	ChannelID           int                         `json:"channel_id"`
	Name                string                      `json:"name"`
	Type                int                         `json:"type"`
	TypeName            string                      `json:"type_name"`
	CreatorUserID       int                         `json:"creator_user_id"`
	CreatorUsername     string                      `json:"creator_username"`
	SupplyRatio         float64                     `json:"supply_ratio"`
	Status              int                         `json:"status"`
	OverallAvailability float64                     `json:"overall_availability"`
	SuccessCount        int                         `json:"success_count"`
	FailureCount        int                         `json:"failure_count"`
	AvgLatencyMs        float64                     `json:"avg_latency_ms"`
	LastTrafficAt       int64                       `json:"last_traffic_at"`
	HasTraffic          bool                        `json:"has_traffic"`
	Minutes             []ChannelAvailabilityMinute `json:"minutes"`
}

type channelAvailabilityAggregateRow struct {
	ChannelID           int
	Name                string
	Type                int
	OwnerUserID         int
	OwnerUsername       string
	SupplyRatio         float64
	Status              int
	SuccessCount        int
	FailureCount        int
	LatencyTotalMs      int64
	LastTrafficAt       int64
}

type channelAvailabilityBucketRow struct {
	ChannelID      int
	BucketStart    int64
	SuccessCount   int
	FailureCount   int
	LatencyTotalMs int64
}

func bucketStartUnix(ts time.Time) int64 {
	return ts.Truncate(time.Minute).Unix()
}

func upsertChannelAvailabilityBucket(channelID int, success bool, latencyMs int64, now time.Time) error {
	if channelID <= 0 {
		return nil
	}
	bucketStart := bucketStartUnix(now)
	updatedAt := now.Unix()

	if common.UsingPostgreSQL {
		successInc := 0
		failureInc := 0
		if success {
			successInc = 1
		} else {
			failureInc = 1
		}
		return DB.Exec(`
			INSERT INTO channel_availability_buckets
				(channel_id, bucket_start, success_count, failure_count, latency_total_ms, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT (channel_id, bucket_start)
			DO UPDATE SET
				success_count = channel_availability_buckets.success_count + EXCLUDED.success_count,
				failure_count = channel_availability_buckets.failure_count + EXCLUDED.failure_count,
				latency_total_ms = channel_availability_buckets.latency_total_ms + EXCLUDED.latency_total_ms,
				updated_at = EXCLUDED.updated_at
		`, channelID, bucketStart, successInc, failureInc, latencyMs, updatedAt).Error
	}

	if common.UsingSQLite {
		successInc := 0
		failureInc := 0
		if success {
			successInc = 1
		} else {
			failureInc = 1
		}
		return DB.Exec(`
			INSERT INTO channel_availability_buckets
				(channel_id, bucket_start, success_count, failure_count, latency_total_ms, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(channel_id, bucket_start)
			DO UPDATE SET
				success_count = success_count + excluded.success_count,
				failure_count = failure_count + excluded.failure_count,
				latency_total_ms = latency_total_ms + excluded.latency_total_ms,
				updated_at = excluded.updated_at
		`, channelID, bucketStart, successInc, failureInc, latencyMs, updatedAt).Error
	}

	successExpr := "0"
	failureExpr := "0"
	if success {
		successExpr = "1"
	} else {
		failureExpr = "1"
	}
	return DB.Exec(`
		INSERT INTO channel_availability_buckets
			(channel_id, bucket_start, success_count, failure_count, latency_total_ms, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			success_count = success_count + `+successExpr+`,
			failure_count = failure_count + `+failureExpr+`,
			latency_total_ms = latency_total_ms + VALUES(latency_total_ms),
			updated_at = VALUES(updated_at)
	`, channelID, bucketStart, boolToInt(success), boolToInt(!success), latencyMs, updatedAt).Error
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func RecordChannelAvailabilitySuccess(channelID int, latencyMs int64) error {
	if latencyMs < 0 {
		latencyMs = 0
	}
	return upsertChannelAvailabilityBucket(channelID, true, latencyMs, time.Now())
}

func RecordChannelAvailabilityFailure(channelID int, latencyMs int64) error {
	if latencyMs < 0 {
		latencyMs = 0
	}
	return upsertChannelAvailabilityBucket(channelID, false, latencyMs, time.Now())
}

func minuteStatus(successCount int, failureCount int) string {
	total := successCount + failureCount
	if total == 0 {
		return "idle"
	}
	if failureCount == 0 {
		return "healthy"
	}
	if successCount == 0 {
		return "failed"
	}
	return "degraded"
}

func GetChannelAvailability(keyword string, startIdx int, num int) ([]ChannelAvailabilityItem, int64, error) {
	if num <= 0 {
		num = common.ItemsPerPage
	}

	windowEnd := time.Now().Truncate(time.Minute)
	windowStart := windowEnd.Add(-time.Duration(channelAvailabilityWindowMinutes-1) * time.Minute)
	windowStartUnix := windowStart.Unix()
	idCastExpr := "CAST(c.id AS CHAR)"
	if common.UsingPostgreSQL || common.UsingSQLite {
		idCastExpr = "CAST(c.id AS TEXT)"
	}

	baseQuery := DB.Table("channels AS c").
		Select(`
			c.id AS channel_id,
			c.name,
			c.type,
			c.owner_user_id,
			COALESCE(u.username, '') AS owner_username,
			c.supply_ratio,
			c.status
		`).
		Joins("LEFT JOIN users AS u ON u.id = c.owner_user_id").
		Where("c.status = ?", common.ChannelStatusEnabled).
		Where("c.type IN ?", constant.SupportedChannelTypeIDs)

	trimmedKeyword := strings.TrimSpace(keyword)
	if trimmedKeyword != "" {
		likeKeyword := "%" + trimmedKeyword + "%"
		baseQuery = baseQuery.Where(
			"c.name LIKE ? OR COALESCE(u.username, '') LIKE ? OR ? = "+idCastExpr,
			likeKeyword,
			likeKeyword,
			trimmedKeyword,
		)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []ChannelAvailabilityItem{}, 0, nil
	}

	aggregateQuery := DB.Table("channels AS c").
		Select(`
			c.id AS channel_id,
			c.name,
			c.type,
			c.owner_user_id,
			COALESCE(u.username, '') AS owner_username,
			c.supply_ratio,
			c.status,
			COALESCE(SUM(b.success_count), 0) AS success_count,
			COALESCE(SUM(b.failure_count), 0) AS failure_count,
			COALESCE(SUM(b.latency_total_ms), 0) AS latency_total_ms,
			COALESCE(MAX(b.updated_at), 0) AS last_traffic_at
		`).
		Joins("LEFT JOIN users AS u ON u.id = c.owner_user_id").
		Joins(
			"LEFT JOIN channel_availability_buckets AS b ON b.channel_id = c.id AND b.bucket_start >= ?",
			windowStartUnix,
		).
		Where("c.status = ?", common.ChannelStatusEnabled).
		Where("c.type IN ?", constant.SupportedChannelTypeIDs)

	if trimmedKeyword != "" {
		likeKeyword := "%" + trimmedKeyword + "%"
		aggregateQuery = aggregateQuery.Where(
			"c.name LIKE ? OR COALESCE(u.username, '') LIKE ? OR ? = "+idCastExpr,
			likeKeyword,
			likeKeyword,
			trimmedKeyword,
		)
	}

	var rows []channelAvailabilityAggregateRow
	if err := aggregateQuery.
		Group("c.id, c.name, c.type, c.owner_user_id, u.username, c.supply_ratio, c.status").
		Order("last_traffic_at DESC, c.id DESC").
		Offset(startIdx).
		Limit(num).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	if len(rows) == 0 {
		return []ChannelAvailabilityItem{}, total, nil
	}

	channelIDs := make([]int, 0, len(rows))
	for _, row := range rows {
		channelIDs = append(channelIDs, row.ChannelID)
	}

	var bucketRows []channelAvailabilityBucketRow
	if err := DB.Table("channel_availability_buckets").
		Select("channel_id, bucket_start, success_count, failure_count, latency_total_ms").
		Where("channel_id IN ?", channelIDs).
		Where("bucket_start >= ?", windowStartUnix).
		Order("bucket_start ASC").
		Scan(&bucketRows).Error; err != nil {
		return nil, 0, err
	}

	bucketsByChannel := make(map[int]map[int64]channelAvailabilityBucketRow, len(channelIDs))
	for _, row := range bucketRows {
		if _, ok := bucketsByChannel[row.ChannelID]; !ok {
			bucketsByChannel[row.ChannelID] = make(map[int64]channelAvailabilityBucketRow)
		}
		bucketsByChannel[row.ChannelID][row.BucketStart] = row
	}

	items := make([]ChannelAvailabilityItem, 0, len(rows))
	for _, row := range rows {
		totalRequests := row.SuccessCount + row.FailureCount
		avgLatency := 0.0
		if totalRequests > 0 {
			avgLatency = float64(row.LatencyTotalMs) / float64(totalRequests)
		}
		overallAvailability := 0.0
		if totalRequests > 0 {
			overallAvailability = float64(row.SuccessCount) / float64(totalRequests)
		}

		minutes := make([]ChannelAvailabilityMinute, 0, channelAvailabilityWindowMinutes)
		for i := 0; i < channelAvailabilityWindowMinutes; i++ {
			bucketTs := windowStart.Add(time.Duration(i) * time.Minute).Unix()
			bucket := bucketsByChannel[row.ChannelID][bucketTs]
			totalBucketRequests := bucket.SuccessCount + bucket.FailureCount
			minuteAvgLatency := 0.0
			minuteAvailability := 0.0
			if totalBucketRequests > 0 {
				minuteAvgLatency = float64(bucket.LatencyTotalMs) / float64(totalBucketRequests)
				minuteAvailability = float64(bucket.SuccessCount) / float64(totalBucketRequests)
			}
			minutes = append(minutes, ChannelAvailabilityMinute{
				BucketStart:   bucketTs,
				SuccessCount:  bucket.SuccessCount,
				FailureCount:  bucket.FailureCount,
				AvgLatencyMs:  minuteAvgLatency,
				Availability:  minuteAvailability,
				Status:        minuteStatus(bucket.SuccessCount, bucket.FailureCount),
				HasTraffic:    totalBucketRequests > 0,
				TotalRequests: totalBucketRequests,
			})
		}

		items = append(items, ChannelAvailabilityItem{
			ChannelID:           row.ChannelID,
			Name:                row.Name,
			Type:                row.Type,
			TypeName:            constant.GetChannelTypeName(row.Type),
			CreatorUserID:       row.OwnerUserID,
			CreatorUsername:     row.OwnerUsername,
			SupplyRatio:         row.SupplyRatio,
			Status:              row.Status,
			OverallAvailability: overallAvailability,
			SuccessCount:        row.SuccessCount,
			FailureCount:        row.FailureCount,
			AvgLatencyMs:        avgLatency,
			LastTrafficAt:       row.LastTrafficAt,
			HasTraffic:          totalRequests > 0,
			Minutes:             minutes,
		})
	}

	return items, total, nil
}

func CleanupOldChannelAvailabilityBuckets(retainDays int) error {
	if retainDays <= 0 {
		retainDays = 7
	}
	cutoff := time.Now().Add(-time.Duration(retainDays) * 24 * time.Hour).Unix()
	if err := DB.Where("bucket_start < ?", cutoff).Delete(&ChannelAvailabilityBucket{}).Error; err != nil {
		return fmt.Errorf("cleanup channel availability buckets failed: %w", err)
	}
	return nil
}
