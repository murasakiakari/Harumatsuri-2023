package service

import (
	"gatewayserver/backend"
	"gatewayserver/utility"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var ticketPool = utility.Pool[Ticket]{
	New: func() any {
		return &Ticket{}
	},
}

type Ticket struct {
	Token string `form:"token" binding:"required"`
}

type SecondEntry struct {
	Quantity int `form:"quantity" binding:"required"`
}

func createTestTicket(ctx *gin.Context) {
	application := &backend.TicketApplication{}
	if err := ctx.ShouldBind(application); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to bind: missing request body"})
		return
	}

	if err := backend.CreateTicket(application); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.Next()
}

func sendConfirmationEmail(ctx *gin.Context) {
	tickerMessage, err := backend.SendConfirmationEmail()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tickerMessage)
}

func sendTicket(ctx *gin.Context) {
	tickerMessage, err := backend.SendTicket()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tickerMessage)
}

func getTicketInfo(ctx *gin.Context) {
	ticket := ticketPool.Get()
	defer ticketPool.Put(ticket)

	if err := ctx.ShouldBind(ticket); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "failed to bind: missing request body"})
		return
	}

	ticketInfo, err := backend.GetTicketInfo(ticket.Token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, ticketInfo)
}

func useTicket(ctx *gin.Context) {
	ticket := ticketPool.Get()
	defer ticketPool.Put(ticket)

	if err := ctx.ShouldBind(ticket); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "failed to bind: missing request body"})
		return
	}

	nTicketUsed_, err := backend.UseTicket(ticket.Token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	nTicketUsed, err := strconv.Atoi(nTicketUsed_)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to convert string to integer: " + err.Error()})
		return
	}

	queuingSystem.UpdateFirstEntry(nTicketUsed)
	ctx.JSON(http.StatusOK, nil)
}

func secondEntry(ctx *gin.Context) {
	secondEntryApplication := &SecondEntry{}
	if err := ctx.ShouldBind(secondEntryApplication); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "failed to bind: missing request body"})
		return
	}

	queuingSystem.UpdateSecondEntry(secondEntryApplication.Quantity)
}
