package controller

import (
	"net/http"
	"server/config"
	"server/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PJSKService = service.PJSKService
type PJSKConfig = config.PJSKConfig

type PJSKController struct{
	r *gin.Engine
	pjskConfig *PJSKConfig
	pjskService *PJSKService
}

func (p *PJSKController) Construct(rP *gin.Engine, cfg  *PJSKConfig, pjskServiceP *PJSKService) {
	p.r = rP
	p.pjskConfig = cfg
	p.pjskService = pjskServiceP
}

func (p *PJSKController) Register() {
	p.getCharts()
	p.getJackets()
	p.r.Run(":" + strconv.Itoa(p.pjskConfig.Server.Port))
}

func (p *PJSKController) getCharts() {
	p.r.GET("/pjsk/charts", func(c *gin.Context) {
		id := c.Query("id")
		level := c.Query("level")
		data, err := p.pjskService.GetCharts(id, level)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Data(http.StatusOK, "image/png", data)
	})
}

func (p *PJSKController) getJackets() {
	p.r.GET("/pjsk/jackets", func(c *gin.Context) {
		id := c.Query("id")
		data, err := p.pjskService.GetJackets(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Data(http.StatusOK, "image/png", data)
	})
}