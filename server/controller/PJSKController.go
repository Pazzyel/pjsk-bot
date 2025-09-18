package controller

import (
	"net/http"
	"server/service"

	"github.com/gin-gonic/gin"
)

type PJSKService = service.PJSKService

type PJSKController struct{
	r *gin.Engine
	pjskService *PJSKService
}

func (p *PJSKController) Construct(rP *gin.Engine, pjskServiceP *PJSKService) {
	p.r = rP
	p.pjskService = pjskServiceP
}

func (p *PJSKController) Register() {
	p.getCharts()
	p.r.Run(":9470")
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