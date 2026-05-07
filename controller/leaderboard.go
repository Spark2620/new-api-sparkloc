package controller

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func GetLeaderboard(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	currentUserId := c.GetInt("id")

	leaderboardType := model.LeaderboardType(strings.TrimSpace(c.Query("type")))
	if leaderboardType != model.LeaderboardTypeContribution {
		leaderboardType = model.LeaderboardTypeConsumption
	}

	items, total, err := model.GetLeaderboardEntries(
		leaderboardType,
		currentUserId,
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, gin.H{
		"type":      leaderboardType,
		"page":      pageInfo.Page,
		"page_size": pageInfo.PageSize,
		"total":     pageInfo.Total,
		"items":     pageInfo.Items,
	})
}
