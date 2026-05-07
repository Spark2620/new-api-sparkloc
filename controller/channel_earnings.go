package controller

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func GetChannelEarnings(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	keyword := strings.TrimSpace(c.Query("keyword"))
	userID := c.GetInt("id")

	items, summary, err := model.GetChannelEarnings(
		userID,
		keyword,
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	total := 0
	if summary != nil {
		total = int(summary.MatchedRequests)
	}
	pageInfo.SetTotal(total)
	pageInfo.SetItems(items)
	common.ApiSuccess(c, gin.H{
		"page":      pageInfo.Page,
		"page_size": pageInfo.PageSize,
		"total":     pageInfo.Total,
		"keyword":   keyword,
		"items":     pageInfo.Items,
		"summary":   summary,
	})
}
