package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type Setup struct {
	Status       bool   `json:"status"`
	RootInit     bool   `json:"root_init"`
	DatabaseType string `json:"database_type"`
}

func GetSetup(c *gin.Context) {
	setup := Setup{
		Status: constant.Setup,
	}
	if !constant.Setup {
		setup.RootInit = model.RootUserExists()
		if common.UsingMySQL {
			setup.DatabaseType = "mysql"
		}
		if common.UsingPostgreSQL {
			setup.DatabaseType = "postgres"
		}
		if common.UsingSQLite {
			setup.DatabaseType = "sqlite"
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    setup,
	})
}

func PostSetup(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": false,
		"message": "系统初始化已改为 Sparkloc OAuth 登录，请使用 Sparkloc 登录完成首次管理员初始化。",
	})
}
