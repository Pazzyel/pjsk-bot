package container

import (
	"server/config"
	"server/controller"
	"server/service"

	"github.com/gin-gonic/gin"
)

type ApplicationContext struct {
	pjskController *controller.PJSKController
	pjskService   *service.PJSKService
	PJSKConfig  *config.PJSKConfig
}

func CreateContext() *ApplicationContext {
	// 加载配置
	cfg, err := config.LoadConfig("resources/config")
	if err != nil {
		panic("加载配置文件失败: " + err.Error())
	}
	if cfg == nil {
		panic("配置文件为空")
	}
	// 初始化 Controller 和 Service
	ps := &service.PJSKService{}
	ps.Construct(cfg)
	pc := &controller.PJSKController{}
	pc.Construct(gin.Default(), cfg , ps)
	// 返回应用容器
	return &ApplicationContext{
		pjskController: pc,
		pjskService:   ps,
		PJSKConfig:  cfg,
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