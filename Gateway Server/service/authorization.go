package service

import (
	"gatewayserver/backend"
	"gatewayserver/utility"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MessageWithRedirect struct {
	Message  string
	Redirect template.URL
}

func login(ctx *gin.Context) {
	account := &backend.Account{}
	if err := ctx.ShouldBind(account); err != nil {
		ctx.Abort()
		ctx.HTML(http.StatusBadRequest, "message-with-redirect.html", MessageWithRedirect{Message: "Missing username or password", Redirect: template.URL("/login")})
		return
	}

	token, err := backend.Login(account)
	if err != nil {
		ctx.Abort()
		ctx.HTML(http.StatusInternalServerError, "message-with-redirect.html", MessageWithRedirect{Message: err.Error(), Redirect: template.URL("/login")})
		return
	}

	ctx.SetCookie("token", token, 3600*10, "/", utility.Config.Domain.Name, true, true)
}

func needLogin(ctx *gin.Context) {
	token, err := ctx.Cookie("token")
	if err != nil {
		ctx.Abort()
		ctx.HTML(http.StatusBadRequest, "message-with-redirect.html", MessageWithRedirect{Message: "Failed to get token", Redirect: template.URL("/login")})
		return
	}

	permission, err := backend.VerifyToken(token)
	if err != nil {
		ctx.Abort()
		ctx.HTML(http.StatusInternalServerError, "message-with-redirect.html", MessageWithRedirect{Message: err.Error(), Redirect: template.URL("/login")})
		return
	}

	ctx.Set("permission", permission)
	ctx.Next()
}

func needPrivilege(ctx *gin.Context) {
	permission_, ok := ctx.Get("permission")
	if !ok {
		ctx.Abort()
		ctx.HTML(http.StatusInternalServerError, "message-with-redirect.html", MessageWithRedirect{Message: "Missing field permission", Redirect: template.URL("/login")})
		return
	}
	permission := permission_.(string)

	if permission != "admin" {
		ctx.Abort()
		ctx.HTML(http.StatusUnauthorized, "message-with-redirect.html", MessageWithRedirect{Message: "No privilege", Redirect: template.URL("/dashboard")})
		return
	}

	ctx.Next()
}

func logout(ctx *gin.Context) {
	ctx.SetCookie("token", "", 0, "/", utility.Config.Domain.Name, true, true)
	ctx.HTML(http.StatusOK, "message-with-redirect.html", MessageWithRedirect{Message: "Logout successfully", Redirect: template.URL("/login")})
}
