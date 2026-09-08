package handler

import (
	"user-center/internal/exception"
	"user-center/internal/middleware"
	"user-center/internal/service"
	"github.com/gin-gonic/gin"
)

type OIDCHandler interface {
	Discovery(c *gin.Context)
	Authorize(c *gin.Context)
	Token(c *gin.Context)
	UserInfo(c *gin.Context)
	JWKS(c *gin.Context)
}

type oidcHandlerImpl struct {
	svc    service.OIDCService
	issuer string
}

func NewOIDCHandler(issuer string) OIDCHandler {
	return &oidcHandlerImpl{
		svc:    service.GetOIDCService(),
		issuer: issuer,
	}
}

// Discovery godoc
// @Summary  OIDC Discovery 元数据
// @Tags     oidc
// @Produce  json
// @Success  200  {object}  map[string]interface{}
// @Router   /.well-known/openid-configuration [get]
func (h *oidcHandlerImpl) Discovery(c *gin.Context) {
	exception.OK(c, h.svc.Discovery(h.issuer))
}

// Authorize godoc
// @Summary  OIDC 授权码端点 (简化版)
// @Tags     oidc
// @Security BearerAuth
// @Produce  json
// @Param    client_id      query  string  true   "客户端 ID"
// @Param    redirect_uri   query  string  true   "回调地址"
// @Param    response_type  query  string  true   "固定为 code"
// @Param    scope          query  string  false  "作用域"
// @Param    state          query  string  false  "状态"
// @Success  200            {object}  exception.APIResponse{data=object{code=string,redirect_uri=string,state=string}}
// @Router   /oidc/authorize [get]
func (h *oidcHandlerImpl) Authorize(c *gin.Context) {
	userID, ok := c.Get(middleware.CtxUserID)
	if !ok {
		exception.Fail(c, exception.NewUnauthorized("authentication required"))
		return
	}
	req := &service.AuthorizeRequest{
		ClientID:     c.Query("client_id"),
		RedirectURI:  c.Query("redirect_uri"),
		ResponseType: c.Query("response_type"),
		Scope:        c.Query("scope"),
		State:        c.Query("state"),
	}
	if req.ClientID == "" || req.RedirectURI == "" || req.ResponseType == "" {
		exception.Fail(c, exception.NewBadRequest("missing required parameters"))
		return
	}
	code, redirect, state, ex := h.svc.Authorize(userID.(string), req)
	if ex != nil {
		exception.Fail(c, ex)
		return
	}
	exception.OK(c, gin.H{"code": code, "redirect_uri": redirect, "state": state})
}

// Token godoc
// @Summary  OIDC Token 端点
// @Tags     oidc
// @Accept   json
// @Produce  json
// @Param    body  body      service.TokenExchangeRequest  true  "授权码/刷新令牌请求"
// @Success  200   {object}  exception.APIResponse{data=service.TokenInfo}
// @Router   /oidc/token [post]
func (h *oidcHandlerImpl) Token(c *gin.Context) {
	var req service.TokenExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		exception.Fail(c, exception.NewBadRequest(err.Error()))
		return
	}
	tok, ex := h.svc.TokenExchange(&req)
	if ex != nil {
		exception.Fail(c, ex)
		return
	}
	exception.OK(c, tok)
}

// UserInfo godoc
// @Summary  OIDC UserInfo 端点
// @Tags     oidc
// @Security BearerAuth
// @Produce  json
// @Success  200  {object}  exception.APIResponse{data=service.UserInfo}
// @Router   /oidc/userinfo [get]
func (h *oidcHandlerImpl) UserInfo(c *gin.Context) {
	token := c.Query("access_token")
	if token == "" {
		raw := c.GetHeader("Authorization")
		if len(raw) > 7 && raw[:7] == "Bearer " {
			token = raw[7:]
		}
	}
	if token == "" {
		exception.Fail(c, exception.NewUnauthorized("missing access token"))
		return
	}
	u, ex := h.svc.UserInfo(token)
	if ex != nil {
		exception.Fail(c, ex)
		return
	}
	exception.OK(c, u)
}

// JWKS godoc
// @Summary  OIDC JWKS 公钥
// @Tags     oidc
// @Produce  json
// @Success  200  {object}  map[string]interface{}
// @Router   /oidc/jwks [get]
func (h *oidcHandlerImpl) JWKS(c *gin.Context) {
	exception.OK(c, h.svc.JWKS())
}
