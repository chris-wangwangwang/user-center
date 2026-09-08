package handler

import (
	"strconv"

	"github.com/chris/usercenter/internal/exception"
	"github.com/chris/usercenter/internal/middleware"
	"github.com/chris/usercenter/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Refresh(c *gin.Context)
	Logout(c *gin.Context)
	Profile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	ChangePassword(c *gin.Context)
	List(c *gin.Context)
	Disable(c *gin.Context)
}

type userHandlerImpl struct {
	svc service.UserService
}

func NewUserHandler() UserHandler {
	return &userHandlerImpl{svc: service.GetUserService()}
}

// Register godoc
// @Summary  用户注册
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body  body      service.RegisterRequest  true  "注册信息"
// @Success  200   {object}  exception.APIResponse{data=service.UserInfo}
// @Router   /api/v1/auth/register [post]
func (h *userHandlerImpl) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		exception.Fail(c, exception.NewBadRequest(err.Error()))
		return
	}
	u, ex := h.svc.Register(&req)
	if ex != nil {
		exception.Fail(c, ex)
		return
	}
	exception.OK(c, u)
}

// Login godoc
// @Summary  用户登录
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body  body      service.LoginRequest  true  "登录凭据"
// @Success  200   {object}  exception.APIResponse{data=service.TokenInfo}
// @Router   /api/v1/auth/login [post]
func (h *userHandlerImpl) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		exception.Fail(c, exception.NewBadRequest(err.Error()))
		return
	}
	tok, ex := h.svc.Login(&req)
	if ex != nil {
		exception.Fail(c, ex)
		return
	}
	exception.OK(c, tok)
}

// Refresh godoc
// @Summary  刷新访问令牌
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body  body      object{refresh_token=string}  true  "刷新令牌"
// @Success  200   {object}  exception.APIResponse{data=service.TokenInfo}
// @Router   /api/v1/auth/refresh [post]
func (h *userHandlerImpl) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		exception.Fail(c, exception.NewBadRequest(err.Error()))
		return
	}
	tok, ex := h.svc.Refresh(req.RefreshToken)
	if ex != nil {
		exception.Fail(c, ex)
		return
	}
	exception.OK(c, tok)
}

// Logout godoc
// @Summary  注销当前会话
// @Tags     auth
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body  body      object{refresh_token=string}  false  "可选: 撤销指定 refresh"
// @Success  200   {object}  exception.APIResponse
// @Router   /api/v1/auth/logout [post]
func (h *userHandlerImpl) Logout(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxUserID)
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&req)
	if ex := h.svc.Logout(userID.(string), req.RefreshToken); ex != nil {
		exception.Fail(c, ex)
		return
	}
	exception.OK(c, nil)
}

// Profile godoc
// @Summary  获取当前用户资料
// @Tags     user
// @Security BearerAuth
// @Produce  json
// @Success  200  {object}  exception.APIResponse{data=service.UserInfo}
// @Router   /api/v1/users/me [get]
func (h *userHandlerImpl) Profile(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxUserID)
	u, ex := h.svc.GetProfile(userID.(string))
	if ex != nil {
		exception.Fail(c, ex)
		return
	}
	exception.OK(c, u)
}

// UpdateProfile godoc
// @Summary  更新当前用户资料
// @Tags     user
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body  body      service.UpdateProfileRequest  true  "更新字段"
// @Success  200   {object}  exception.APIResponse{data=service.UserInfo}
// @Router   /api/v1/users/me [put]
func (h *userHandlerImpl) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxUserID)
	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		exception.Fail(c, exception.NewBadRequest(err.Error()))
		return
	}
	u, ex := h.svc.UpdateProfile(userID.(string), &req)
	if ex != nil {
		exception.Fail(c, ex)
		return
	}
	exception.OK(c, u)
}

// ChangePassword godoc
// @Summary  修改当前用户密码
// @Tags     user
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body  body      service.ChangePasswordRequest  true  "旧密码与新密码"
// @Success  200   {object}  exception.APIResponse
// @Router   /api/v1/users/me/password [post]
func (h *userHandlerImpl) ChangePassword(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxUserID)
	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		exception.Fail(c, exception.NewBadRequest(err.Error()))
		return
	}
	if ex := h.svc.ChangePassword(userID.(string), &req); ex != nil {
		exception.Fail(c, ex)
		return
	}
	exception.OK(c, nil)
}

// List godoc
// @Summary  分页查询用户列表
// @Tags     user
// @Security BearerAuth
// @Produce  json
// @Param    page  query     int  false  "页码"
// @Param    size  query     int  false  "每页数量"
// @Success  200   {object}  exception.APIResponse{data=object{list=[]service.UserInfo,total=int64}}
// @Router   /api/v1/users [get]
func (h *userHandlerImpl) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	users, total, ex := h.svc.List(page, size)
	if ex != nil {
		exception.Fail(c, ex)
		return
	}
	exception.OK(c, gin.H{"list": users, "total": total})
}

// Disable godoc
// @Summary  启用/禁用用户
// @Tags     user
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id    path      string  true   "用户 ID"
// @Param    body  body      object{disabled=bool}  true  "禁用状态"
// @Success  200   {object}  exception.APIResponse
// @Router   /api/v1/users/{id}/disable [post]
func (h *userHandlerImpl) Disable(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Disabled bool `json:"disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		exception.Fail(c, exception.NewBadRequest(err.Error()))
		return
	}
	if ex := h.svc.Disable(id, req.Disabled); ex != nil {
		exception.Fail(c, ex)
		return
	}
	exception.OK(c, nil)
}
