package service

import (
	"backendserver/database"
	"backendserver/utility"
	"fmt"
	"io"
	"text/template"

	"github.com/golang-jwt/jwt/v5"
)

var ticketPool = utility.Pool[database.TicketInfo]{
	New: func() any {
		return &database.TicketInfo{}
	},
}

var (
	vipTicketTemplate      = template.Must(template.New("vip-ticket.html").Funcs(funcMap).ParseFiles("./resources/vip-ticket.html"))
	regularTicketTemplate  = template.Must(template.New("regular-ticket.html").Funcs(funcMap).ParseFiles("./resources/regular-ticket.html"))
	discountTicketTemplate = template.Must(template.New("discount-ticket.html").Funcs(funcMap).ParseFiles("./resources/discount-ticket.html"))

	statusList = []string{"unused", "used"}
)

type Ticket interface {
	Send() error
}

type TicketInfo struct {
	database.TicketInfo
	Embed *utility.EmbedFile
}

type VipTicket struct {
	VipApplication
	Tickets []*TicketInfo
}

func (ticket *VipTicket) Send() (err error) {
	if ticket.Valid == database.INVALID {
		return fmt.Errorf("failed to send ticket for %v: application invalid", ticket.Name)
	}

	ticket.Status = database.TICKET_SENDED
	if err = ticket.Update(); err != nil {
		return fmt.Errorf("failed to update VIP application status: %w", err)
	}

	embedFiles := make([]*utility.EmbedFile, ticket.Quantity)
	for i, ticket := range ticket.Tickets {
		embedFiles[i] = ticket.Embed
	}

	if err = utility.SendEmail(ticket.Email, "[城大春祭2023]VIP門票", func(w io.Writer) error { return vipTicketTemplate.Execute(w, ticket) }, embedFiles...); err == nil {
		return nil
	}
	err = fmt.Errorf("failed to send VIP ticket for %v: %w", ticket.Name, err)

	errs := utility.Errors{err}
	if rollbackErr := ticket.rollback(database.EMAIL_SENDED); rollbackErr != nil {
		errs = append(errs, rollbackErr)
	}
	
	for _, ticketInfo := range ticket.Tickets {
		if deleteErr := ticketInfo.TicketInfo.Delete(); deleteErr != nil {
			errs = append(errs, deleteErr)
		}
	}

	return errs
}

type RegularTicket struct {
	RegularApplication
	Ticket *TicketInfo
}

func (ticket *RegularTicket) Send() (err error) {
	if ticket.Valid == database.INVALID {
		return fmt.Errorf("failed to send Regular ticket for %v: application invalid", ticket.Name)
	}

	ticket.Status = database.TICKET_SENDED
	if err = ticket.Update(); err != nil {
		return fmt.Errorf("failed to update Regular application status: %w", err)
	}

	if err = utility.SendEmail(ticket.Email, "[城大春祭2023]普通門票換領通知", func(w io.Writer) error { return regularTicketTemplate.Execute(w, ticket) }, ticket.Ticket.Embed); err == nil {
		return nil
	}
	err = fmt.Errorf("failed to send Regular ticket for %v: %w", ticket.Name, err)

	errs := utility.Errors{err}
	if rollbackErr := ticket.rollback(database.EMAIL_SENDED); rollbackErr != nil {
		errs = append(errs, rollbackErr)
	}

	if deleteErr := ticket.Ticket.TicketInfo.Delete(); deleteErr != nil {
		errs = append(errs, deleteErr)
	}
	return errs
}

type DiscountTicket struct {
	DiscountApplication
	Ticket *TicketInfo
}

func (ticket *DiscountTicket) Send() (err error) {
	if ticket.Valid == database.INVALID {
		return fmt.Errorf("failed to send Discount ticket for %v: application invalid", ticket.Name)
	}

	ticket.Status = database.TICKET_SENDED
	if err = ticket.Update(); err != nil {
		return fmt.Errorf("failed to update Discount application status: %w", err)
	}

	if err = utility.SendEmail(ticket.Email, "[城大春祭2023]優惠門票換領通知", func(w io.Writer) error { return discountTicketTemplate.Execute(w, ticket) }, ticket.Ticket.Embed); err == nil {
		return nil
	}
	err = fmt.Errorf("failed to send Discount ticket for %v: %w", ticket.Name, err)

	errs := utility.Errors{err}
	if rollbackErr := ticket.rollback(database.EMAIL_SENDED); rollbackErr != nil {
		return utility.Errors{err, rollbackErr}
	}

	if deleteErr := ticket.Ticket.TicketInfo.Delete(); deleteErr != nil {
		errs = append(errs, deleteErr)
	}
	return errs
}

func sendVipTicket() (sent int, err error) {
	applications_, err := database.VipApplications{}.GetApplicationWithStatus(database.EMAIL_SENDED)
	if err != nil {
		return 0, fmt.Errorf("failed to get VIP application: %w", err)
	}

	applications := make([]*VipApplication, len(applications_))
	for i := range applications_ {
		applications[i] = &VipApplication{*applications_[i]}
	}

	sent, err = sentTicket(applications...)
	if err != nil {
		return sent, fmt.Errorf("failed to send VIP ticket: %w", err)
	}
	return sent, nil
}

func sendRegularTicket() (sent int, err error) {
	applications_, err := database.RegularApplications{}.GetApplicationWithStatus(database.EMAIL_SENDED)
	if err != nil {
		return 0, fmt.Errorf("failed to get Regular application: %w", err)
	}

	applications := make([]*RegularApplication, len(applications_))
	for i := range applications_ {
		applications[i] = &RegularApplication{*applications_[i]}
	}

	sent, err = sentTicket(applications...)
	if err != nil {
		return sent, fmt.Errorf("failed to send Regular ticket: %w", err)
	}
	return sent, nil
}

func sendDiscountTicket() (sent int, err error) {
	applications_, err := database.DiscountApplications{}.GetApplicationWithStatus(database.EMAIL_SENDED)
	if err != nil {
		return 0, fmt.Errorf("failed to get Discount application: %w", err)
	}

	applications := make([]*DiscountApplication, len(applications_))
	for i := range applications_ {
		applications[i] = &DiscountApplication{*applications_[i]}
	}

	sent, err = sentTicket(applications...)
	if err != nil {
		return sent, fmt.Errorf("failed to send Discount ticket: %w", err)
	}
	return sent, nil
}

func sentTicket[T Application](applications ...T) (sent int, err error) {
	errs := utility.Errors{}
	for _, application := range applications {
		ticket, err := application.GenerateTicket()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if err = ticket.Send(); err != nil {
			errs = append(errs, err)
			continue
		}
		sent++
	}

	if errs.HasError() {
		return sent, errs
	}
	return sent, nil
}

func getTicketInfo(token string) (claims *jwt.RegisteredClaims, Status string, err error) {
	ticketInfo := ticketPool.Get()
	defer ticketPool.Put(ticketInfo)

	ticketInfo.Token = token
	if err := ticketInfo.Read(); err != nil {
		return nil, "", err
	}

	claims, err = ticketInfo.Validate()
	if err != nil {
		return nil, "", err
	}
	return claims, statusList[ticketInfo.Status], nil
}

func useTicket(token string) (claims *jwt.RegisteredClaims, err error) {
	ticketInfo := ticketPool.Get()
	defer ticketPool.Put(ticketInfo)

	ticketInfo.Token = token
	if err := ticketInfo.Read(); err != nil {
		return nil, err
	}

	claims, err = ticketInfo.Validate()
	if err != nil {
		return nil, err
	}

	ticketInfo.Status = database.USED
	if err := ticketInfo.Update(); err != nil {
		return nil, err
	}

	return claims, nil
}
