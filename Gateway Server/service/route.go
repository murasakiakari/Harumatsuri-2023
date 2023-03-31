package service

import (
	"gatewayserver/utility"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/murasakiakari/pathlib"
)

func New() (err error) {
	logPath := pathlib.CurrentExecutablePath.Dir().Join(utility.Config.Log.FileName)
	flatDatabaseWriter, err := utility.NewFlatDatabaseWriter(logPath.String())
	if err != nil {
		return err
	}
	defer flatDatabaseWriter.Close()

	gin.SetMode(utility.Config.Server.Mode)
	gin.DisableConsoleColor()
	gin.DefaultWriter = io.MultiWriter(flatDatabaseWriter, os.Stdout)
	server := gin.Default()
	server.TrustedPlatform = gin.PlatformCloudflare

	server.LoadHTMLGlob("./resources/*.html")
	server.StaticFile("/resources/style.css", "./resources/style.css")
	server.GET("/login", func(ctx *gin.Context) { ctx.HTML(http.StatusOK, "login.html", nil) })
	server.POST("/login", login, func(ctx *gin.Context) { ctx.Redirect(http.StatusFound, "/dashboard") })
	server.GET("/logout", logout)
	server.GET("/dashboard", needLogin, func(ctx *gin.Context) {
		permission_, ok := ctx.Get("permission")
		if !ok {
			ctx.HTML(http.StatusInternalServerError, "message-with-redirect.html", MessageWithRedirect{Message: "Missing field permission", Redirect: template.URL("/login")})
			return
		}

		switch permission_.(string) {
		case "admin":
			ctx.HTML(http.StatusOK, "dashboard.html", nil)
		case "user":
			ctx.HTML(http.StatusOK, "dashboard-user.html", nil)
		}
	})
	server.GET("/queuing-panel", needLogin, needPrivilege, func(ctx *gin.Context) { ctx.HTML(http.StatusOK, "queuing-panel.html", nil) })
	server.GET("/ticket", needLogin, needPrivilege, func(ctx *gin.Context) { ctx.HTML(http.StatusOK, "ticket.html", nil) })
	server.POST("/ticket", needLogin, needPrivilege, createTestTicket, func(ctx *gin.Context) { ctx.Redirect(http.StatusFound, "/ticket") })
	server.GET("/ticket-scanner", needLogin, func(ctx *gin.Context) {
		permission_, ok := ctx.Get("permission")
		if !ok {
			ctx.HTML(http.StatusInternalServerError, "message-with-redirect.html", MessageWithRedirect{Message: "Missing field permission", Redirect: template.URL("/login")})
			return
		}

		switch permission_.(string) {
		case "admin":
			ctx.HTML(http.StatusOK, "ticket-scanner.html", nil)
		case "user":
			ctx.HTML(http.StatusOK, "ticket-scanner-user.html", nil)
		}
	})
	server.GET("/second-entry", needLogin, func(ctx *gin.Context) {
		permission_, ok := ctx.Get("permission")
		if !ok {
			ctx.HTML(http.StatusInternalServerError, "message-with-redirect.html", MessageWithRedirect{Message: "Missing field permission", Redirect: template.URL("/login")})
			return
		}

		switch permission_.(string) {
		case "admin":
			ctx.HTML(http.StatusOK, "second-entry.html", nil)
		case "user":
			ctx.HTML(http.StatusOK, "second-entry-user.html", nil)
		}
	})
	server.POST("/second-entry", needLogin, secondEntry, func(ctx *gin.Context) { ctx.Redirect(http.StatusFound, "/second-entry") })
	server.GET("/early-leave", needLogin, func(ctx *gin.Context) { ctx.HTML(http.StatusOK, "early-leave.html", nil) })
	server.GET("/queuing-internal", needLogin, func(ctx *gin.Context) { ctx.HTML(http.StatusOK, "queuing-internal.html", nil) })
	server.GET("/queuing", func(ctx *gin.Context) { ctx.HTML(http.StatusOK, "queuing.html", utility.Config.Domain.Name) })

	api := server.Group("/api")
	api.POST("/start-queuing", needLogin, needPrivilege, queuingSystem.StartQueuing)
	api.POST("/allow-entry", needLogin, needPrivilege, queuingSystem.AllowEntry)
	api.POST("/offset", needLogin, needPrivilege, queuingSystem.SetOffset)
	api.GET("/send-confirmation-email", needLogin, needPrivilege, sendConfirmationEmail)
	api.GET("/send-ticket", needLogin, needPrivilege, sendTicket)
	api.POST("/ticket-status", needLogin, getTicketInfo)
	api.PUT("/ticket-status", needLogin, useTicket)
	api.GET("/queuing-number", queuingSystem.ConnectClient)
	api.GET("/queuing-number/:number", queuingSystem.GetIndex)

	return server.RunTLS(utility.Config.TLS.Port, utility.Config.TLS.CertFile, utility.Config.TLS.KeyFile)
}
