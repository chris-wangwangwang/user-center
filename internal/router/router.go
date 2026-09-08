package router

import (
	"user-center/internal/handler"
	"user-center/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(issuer string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	userH := handler.NewUserHandler()
	oidcH := handler.NewOIDCHandler(issuer)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/.well-known/openid-configuration", adapter(oidcH.Discovery))

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", adapter(userH.Register))
			auth.POST("/login", adapter(userH.Login))
			auth.POST("/refresh", adapter(userH.Refresh))
		}

		secured := v1.Group("")
		secured.Use(middleware.AuthRequired())
		{
			secured.POST("/auth/logout", adapter(userH.Logout))
			secured.GET("/users/me", adapter(userH.Profile))
			secured.PUT("/users/me", adapter(userH.UpdateProfile))
			secured.POST("/users/me/password", adapter(userH.ChangePassword))
			secured.GET("/users", adapter(userH.List))
			secured.POST("/users/:id/disable", adapter(userH.Disable))
		}
	}

	oidc := r.Group("/oidc")
	{
		oidc.GET("/discovery", adapter(oidcH.Discovery))
		oidc.GET("/jwks", adapter(oidcH.JWKS))
		oidc.POST("/token", adapter(oidcH.Token))
		oidc.GET("/userinfo", adapter(oidcH.UserInfo))
		secured := oidc.Group("")
		secured.Use(middleware.AuthRequired())
		secured.GET("/authorize", adapter(oidcH.Authorize))
	}

	return r
}

type ginHandler = gin.HandlerFunc

func adapter(fn ginHandler) gin.HandlerFunc { return fn }
