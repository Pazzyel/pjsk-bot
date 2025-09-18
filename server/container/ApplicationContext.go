package container

import (
	"server/controller"
	"server/service"

	"github.com/gin-gonic/gin"
)

type ApplicationContext struct {
	pjskController *controller.PJSKController
	pjskService   *service.PJSKService
}

func CreateContext() *ApplicationContext {
	ps := &service.PJSKService{}
	pc := &controller.PJSKController{}
	pc.Construct(gin.Default(), ps)
	return &ApplicationContext{
		pjskController: pc,
		pjskService:   ps,
	}
}

func (ctx *ApplicationContext) GetPJSKController() *controller.PJSKController {
	return ctx.pjskController
}

func (ctx *ApplicationContext) GetPJSKService() *service.PJSKService {
	return ctx.pjskService
}