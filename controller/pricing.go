package controller

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
)

type PricingChannelGroup struct {
	Group         string  `json:"group"`
	ChannelId     int     `json:"channel_id"`
	Name          string  `json:"name"`
	OwnerUserId   int     `json:"owner_user_id"`
	OwnerUsername string  `json:"owner_username,omitempty"`
	SupplyRatio   float64 `json:"supply_ratio"`
	Tag           string  `json:"tag,omitempty"`
	Type          int     `json:"type"`
}

func filterPricingByUsableGroups(pricing []model.Pricing, usableGroup map[string]string) []model.Pricing {
	if len(pricing) == 0 {
		return pricing
	}
	if len(usableGroup) == 0 {
		return []model.Pricing{}
	}

	filtered := make([]model.Pricing, 0, len(pricing))
	for _, item := range pricing {
		if common.StringsContains(item.EnableGroup, "all") {
			filtered = append(filtered, item)
			continue
		}
		for _, group := range item.EnableGroup {
			if _, ok := usableGroup[group]; ok {
				filtered = append(filtered, item)
				break
			}
		}
	}
	return filtered
}

func filterPricingByChannelModelGroups(pricing []model.Pricing, modelGroups map[string][]string) []model.Pricing {
	if len(pricing) == 0 || len(modelGroups) == 0 {
		return []model.Pricing{}
	}

	filtered := make([]model.Pricing, 0, len(pricing))
	for _, item := range pricing {
		groups := modelGroups[item.ModelName]
		if len(groups) == 0 {
			continue
		}
		item.EnableGroup = groups
		filtered = append(filtered, item)
	}
	return filtered
}

func collectPricingChannelGroups() (map[string]PricingChannelGroup, map[string][]string) {
	channels, err := model.GetAllChannels(0, 0, true, false)
	if err != nil {
		common.SysLog(fmt.Sprintf("failed to load pricing channel groups: %v", err))
		return map[string]PricingChannelGroup{}, map[string][]string{}
	}

	channelGroups := make(map[string]PricingChannelGroup, len(channels))
	modelGroupSets := make(map[string]map[string]struct{})
	for _, channel := range channels {
		if channel == nil || channel.Status != common.ChannelStatusEnabled || !constant.IsSupportedChannelType(channel.Type) {
			continue
		}
		group := model.ChannelGroupName(channel.Id)
		supplyRatio := channel.SupplyRatio
		if supplyRatio <= 0 {
			supplyRatio = 1
		}
		ownerUsername := channel.OwnerUsername
		if ownerUsername == "" && channel.OwnerUserId > 0 {
			if username, err := model.GetUsernameById(channel.OwnerUserId, false); err == nil {
				ownerUsername = username
			}
		}
		channelGroups[group] = PricingChannelGroup{
			Group:         group,
			ChannelId:     channel.Id,
			Name:          channel.Name,
			OwnerUserId:   channel.OwnerUserId,
			OwnerUsername: ownerUsername,
			SupplyRatio:   supplyRatio,
			Tag:           channel.GetTag(),
			Type:          channel.Type,
		}
		for _, modelName := range channel.GetModels() {
			modelName = strings.TrimSpace(modelName)
			if modelName == "" {
				continue
			}
			groups := modelGroupSets[modelName]
			if groups == nil {
				groups = make(map[string]struct{})
				modelGroupSets[modelName] = groups
			}
			groups[group] = struct{}{}
		}
	}

	modelGroups := make(map[string][]string, len(modelGroupSets))
	for modelName, set := range modelGroupSets {
		groups := make([]string, 0, len(set))
		for group := range set {
			groups = append(groups, group)
		}
		modelGroups[modelName] = groups
	}

	return channelGroups, modelGroups
}

func GetPricing(c *gin.Context) {
	pricing := model.GetPricing()
	usableGroup := map[string]string{}
	groupRatio := map[string]float64{}
	channelGroups, modelGroups := collectPricingChannelGroups()
	for group, channelGroup := range channelGroups {
		usableGroup[group] = channelGroup.Name
		groupRatio[group] = channelGroup.SupplyRatio
	}
	pricing = filterPricingByChannelModelGroups(pricing, modelGroups)

	c.JSON(200, gin.H{
		"success":            true,
		"data":               pricing,
		"vendors":            model.GetVendors(),
		"group_ratio":        groupRatio,
		"usable_group":       usableGroup,
		"channel_groups":     channelGroups,
		"supported_endpoint": model.GetSupportedEndpointMap(),
		"auto_groups":        []string{},
		"pricing_version":    "a42d372ccf0b5dd13ecf71203521f9d2",
	})
}

func ResetModelRatio(c *gin.Context) {
	defaultStr := ratio_setting.DefaultModelRatio2JSONString()
	err := model.UpdateOption("ModelRatio", defaultStr)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = ratio_setting.UpdateModelRatioByJSONString(defaultStr)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "重置模型倍率成功",
	})
}
