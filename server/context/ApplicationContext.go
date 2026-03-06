package context

import (
	"pjsk-bot/server/config"
	"pjsk-bot/server/controller"
	"pjsk-bot/server/service"

	"github.com/gin-gonic/gin"
)

type ApplicationContext struct {
	controllers    []controller.Controller
	pjskController *controller.PJSKController
	pjskService    *service.PJSKService
	PJSKConfig     *config.PJSKConfig
}

func NewContext() *ApplicationContext {
	// 加载配置
	cfg, err := config.LoadConfig("resources/config")
	if err != nil {
		panic("加载配置文件失败: " + err.Error())
	}
	if cfg == nil {
		panic("配置文件为空")
	}
	// 初始化 Controller 和 Service
	ps := service.NewPJSKService(cfg)
	pc := controller.NewPJSKController(gin.Default(), cfg, ps)

	context := &ApplicationContext{
		pjskService: ps,
		PJSKConfig:  cfg,
	}

	controllers := make([]controller.Controller, 0)
	controllers = append(controllers, pc)
	context.controllers = controllers

	// 返回应用容器
	return context
}

func (ctx *ApplicationContext) Run() {
	for _, ctlr := range ctx.controllers {
		ctlr.Register()
	}
}

func (ctx *ApplicationContext) GetPJSKController() *controller.PJSKController {
	return ctx.pjskController
}

func (ctx *ApplicationContext) GetPJSKService() *service.PJSKService {
	return ctx.pjskService
}

func (ctx *ApplicationContext) GetPJSKConfig() *config.PJSKConfig {
	return ctx.PJSKConfig
}
