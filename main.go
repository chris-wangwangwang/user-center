// Package main is the entrypoint of the usercenter OIDC service.
//
// @title           UserCenter OIDC Service
// @version         1.0
// @description     用户中心与 OIDC 服务, 提供注册/登录/鉴权等接口, 供其他分布式服务做统一身份认证。
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
package main

import (
	"fmt"
	"log"

	_ "github.com/chris/usercenter/docs"
	"github.com/chris/usercenter/internal/config"
	"github.com/chris/usercenter/internal/database"
	"github.com/chris/usercenter/internal/router"
)

func main() {
	cfg := config.Get()
	database.Init()

	issuer := fmt.Sprintf("http://localhost:%d", cfg.HTTPPort)
	r := router.New(issuer)

	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	log.Printf("usercenter listening on %s, issuer=%s", addr, issuer)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
