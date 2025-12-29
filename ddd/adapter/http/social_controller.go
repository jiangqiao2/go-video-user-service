package http

import (
	"sync"
	"user-service/ddd/application/app"
	"user-service/ddd/application/cqe"
	"user-service/pkg/assert"
	"user-service/pkg/authctx"
	"user-service/pkg/errno"
	"user-service/pkg/manager"
	"user-service/pkg/middleware"
	"user-service/pkg/restapi"

	"github.com/gin-gonic/gin"
)

var (
	socialControllerOnce      sync.Once
	singletonSocialController SocialController
)

type SocialControllerPlugin struct{}

func (p *SocialControllerPlugin) Name() string {
	return "socialControllerPlugin"
}

func (p *SocialControllerPlugin) MustCreateController() manager.Controller {
	assert.NotCircular()
	socialControllerOnce.Do(func() {
		singletonSocialController = &socialControllerImpl{
			socialApp: app.DefaultSocialApp(),
		}
	})
	assert.NotNil(singletonSocialController)
	return singletonSocialController
}

type SocialController interface {
	manager.Controller
	ToggleFollow(ctx *gin.Context)
	GetUserRelation(ctx *gin.Context)
}

type socialControllerImpl struct {
	manager.Controller
	socialApp app.SocialApp
}

func (c *socialControllerImpl) RegisterOpenApi(router *gin.RouterGroup) {
	v1 := router.Group("user/v1/open/relation")
	{
		v1.GET("", middleware.AuthOptional(), c.GetUserRelation)
	}
}

func (c *socialControllerImpl) RegisterInnerApi(router *gin.RouterGroup) {
	v1 := router.Group("user/v1/inner/relation")
	{
		v1.POST("/follow/toggle", middleware.AuthRequired(), c.ToggleFollow)
	}
}

func (c *socialControllerImpl) RegisterDebugApi(router *gin.RouterGroup) {}
func (c *socialControllerImpl) RegisterOpsApi(router *gin.RouterGroup)   {}

// ToggleFollow 统一的关注/取关接口，通过 follow 字段控制是关注还是取消关注。
func (c *socialControllerImpl) ToggleFollow(ctx *gin.Context) {
	userUUID, err := authctx.MustGetUserUUID(ctx)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	var req cqe.FollowToggleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		restapi.Failed(ctx, errno.NewSimpleBizError(errno.ErrParameterInvalid, err, "target_uuid"))
		return
	}
	req.UserUUID = userUUID
	if err := c.socialApp.ToggleFollow(ctx.Request.Context(), &req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, nil)
}

func (c *socialControllerImpl) GetUserRelation(ctx *gin.Context) {
	var req cqe.CheckFollowReq
	userUUID, _ := authctx.MustGetUserUUID(ctx)
	if err := ctx.ShouldBindQuery(&req); err != nil {
		restapi.Failed(ctx, err)
		return
	}
	req.FollowerUUID = userUUID
	res, err := c.socialApp.GetUserRelationStat(ctx.Request.Context(), &req)
	if err != nil {
		restapi.Failed(ctx, err)
		return
	}
	restapi.Success(ctx, res)
}
