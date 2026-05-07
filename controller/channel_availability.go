package controller

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func GetChannelAvailability(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	keyword := strings.TrimSpace(c.Query("keyword"))

	items, total, err := model.GetChannelAvailability(
		keyword,
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
		"page":      pageInfo.Page,
		"page_size": pageInfo.PageSize,
		"total":     pageInfo.Total,
		"keyword":   keyword,
		"items":     pageInfo.Items,
		"window": gin.H{
			"minutes": 60,
		},
	})
}
