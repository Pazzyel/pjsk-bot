package controller

import (
	"fmt"
	"log"
	"net/http"
	"pjsk-bot/server/config"
	"pjsk-bot/server/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PJSKService = service.PJSKService
type PJSKConfig = config.PJSKConfig

type PJSKController struct {
	r           *gin.Engine
	pjskConfig  *PJSKConfig
	pjskService *PJSKService
}

func NewPJSKController(rP *gin.Engine, cfg *PJSKConfig, pjskServiceP *PJSKService) *PJSKController {
	p := &PJSKController{}
	p.r = rP
	p.pjskConfig = cfg
	p.pjskService = pjskServiceP
	return p
}

func (p *PJSKController) Register() {
	p.getCharts()
	p.getJackets()
	p.updateInfos()
	p.searchInfos()
	p.r.Run(":" + strconv.Itoa(p.pjskConfig.Server.Port))
}

func (p *PJSKController) getCharts() {
	p.r.GET("/pjsk/charts", func(c *gin.Context) {
		id := c.Query("id")
		if len(id) < 4 {
			idInt, _ := strconv.Atoi(id)
			id = fmt.Sprintf("%04d", idInt)
		}
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
		if len(id) < 3 {
			idInt, _ := strconv.Atoi(id)
			id = fmt.Sprintf("%03d", idInt)
		}
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

func (p *PJSKController) updateInfos() {
	p.r.POST("/pjsk/update", func(c *gin.Context) {
		count, err := p.pjskService.UpdateMusicInfos()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"count": count,
		})
	})
}

func (p *PJSKController) searchInfos() {
	p.r.GET("/pjsk/search", func(c *gin.Context) {
		var req SearchRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		log.Println(req)

		result, total, err := p.pjskService.SearchMusicInfos(service.SearchOptions{
			Title:    req.Title,
			Author:   req.Author,
			MinLevel: req.MinLevel,
			MaxLevel: req.MaxLevel,
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"total": total,
			"data":  result,
		})
	})
}
