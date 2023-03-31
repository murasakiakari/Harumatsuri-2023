package service

import (
	"backendserver/database"
	"backendserver/utility"
	"bytes"
	"fmt"
	"image/color"
	"io"
	"strconv"
	"text/template"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	_HARUMATSURI     string = "CityU 2023 Harumatsuri"
	_VIP_TICKET      string = "VIP Ticket"
	_REGULAR_TICKET  string = "Regular Ticket"
	_DISCOUNT_TICKET string = "Discount Ticket"
)

var (
	_PURPLE = color.RGBA{R: 0x3C, G: 0x13, B: 0x61, A: 0xFF}
	_GREEN  = color.RGBA{R: 0x05, G: 0x66, B: 0x08, A: 0xFF}
	_RED    = color.RGBA{R: 0x8B, G: 0x00, B: 0x00, A: 0xFF}

	funcMap = template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}
)

var (
	vipConfirmationEmailTemplate      = template.Must(template.New("vip-confirmation-email.html").Funcs(funcMap).ParseFiles("./resources/vip-confirmation-email.html"))
	vipNotificationEmailTemplate      = template.Must(template.New("vip-notification-email.html").Funcs(funcMap).ParseFiles("./resources/vip-notification-email.html"))
	regularConfirmationEmailTemplate  = template.Must(template.New("regular-confirmation-email.html").Funcs(funcMap).ParseFiles("./resources/regular-confirmation-email.html"))
	regularNotificationEmailTemplate  = template.Must(template.New("regular-notification-email.html").Funcs(funcMap).ParseFiles("./resources/regular-notification-email.html"))
	discountConfirmationEmailTemplate = template.Must(template.New("discount-confirmation-email.html").Funcs(funcMap).ParseFiles("./resources/discount-confirmation-email.html"))
	discountNotificationEmailTemplate = template.Must(template.New("discount-notification-email.html").Funcs(funcMap).ParseFiles("./resources/discount-notification-email.html"))
)

type Application interface {
	SendEmail() error
	GenerateTicket() (Ticket, error)
}

type VipApplication struct {
	database.VipApplication
}

func (application *VipApplication) rollback(status database.ApplicationStatus) error {
	application.Status = status
	if err := application.Update(); err != nil {
		return fmt.Errorf("failed to rollback VIP application status for %v: %w", application.Name, err)
	}
	return nil
}

func (application *VipApplication) SendEmail() (err error) {
	application.Status = database.EMAIL_SENDED
	if err := application.Update(); err != nil {
		return fmt.Errorf("failed to update VIP application status for %v: %w", application.Name, err)
	}

	switch application.Valid {
	case database.VALID:
		if err = utility.SendEmail(application.Email, "[城大春祭2023]VIP門票確認電郵", func(w io.Writer) error { return vipConfirmationEmailTemplate.Execute(w, application) }); err == nil {
			return nil
		}
		err = fmt.Errorf("failed to send VIP confirmation email for %v: %w", application.Name, err)
	case database.INVALID:
		if err = utility.SendEmail(application.Email, "[城大春祭2023]VIP門票付款無效通知電郵", func(w io.Writer) error { return vipNotificationEmailTemplate.Execute(w, application) }); err == nil {
			return nil
		}
		err = fmt.Errorf("failed to send VIP notification email for %v: %w", application.Name, err)
	default:
		err = fmt.Errorf("failed to send VIP notification email for %v: status invalid", application.Name)
	}

	if rollbackErr := application.rollback(database.RECEIVE); rollbackErr != nil {
		return utility.Errors{err, rollbackErr}
	}
	return err
}

func (application *VipApplication) GenerateTicket() (Ticket, error) {
	if application.Valid == database.INVALID {
		return nil, fmt.Errorf("failed to generate ticket for %v: application invalid", application.Name)
	}

	tickets := make([]*TicketInfo, application.Quantity)
	for i := 0; i < application.Quantity; i++ {
		currentTime := time.Now()
		claims := jwt.RegisteredClaims{
			Issuer:   _HARUMATSURI,
			Subject:  _VIP_TICKET,
			Audience: jwt.ClaimStrings{application.Name},
			IssuedAt: jwt.NewNumericDate(currentTime),
			ID:       strconv.Itoa(i),
		}

		token, err := utility.SignClaims(claims)
		if err != nil {
			return nil, fmt.Errorf("failed to sign claims for %v: %w", application.Name, err)
		}

		ticketInfo := &database.TicketInfo{
			Token:  token,
			Name:   application.Name,
			Key:    utility.SecretKey,
			Status: database.UNUSED,
		}

		if err := ticketInfo.Create(); err != nil {
			return nil, fmt.Errorf("failed to create ticket for %v: %w", application.Name, err)
		}

		qrcode, err := utility.GenerateQRCode(token, color.White, _PURPLE)
		if err != nil {
			return nil, fmt.Errorf("failed to generate QR code for %v, %w", application.Name, err)
		}

		embed := &utility.EmbedFile{
			Name:   fmt.Sprintf("%v.png", utility.HashSha3(qrcode)[:16]),
			Reader: bytes.NewBuffer(qrcode),
		}

		tickets[i] = &TicketInfo{
			TicketInfo: *ticketInfo,
			Embed:      embed,
		}
	}

	vipTicket := &VipTicket{
		VipApplication: *application,
		Tickets:        tickets,
	}

	return vipTicket, nil
}

type RegularApplication struct {
	database.RegularApplication
}

func (application *RegularApplication) rollback(status database.ApplicationStatus) error {
	application.Status = status
	if err := application.Update(); err != nil {
		return fmt.Errorf("failed to rollback Regular application status for %v: %w", application.Name, err)
	}
	return nil
}

func (application *RegularApplication) SendEmail() (err error) {
	application.Status = database.EMAIL_SENDED
	if err = application.Update(); err != nil {
		return fmt.Errorf("failed to update Regular application status for %v: %w", application.Name, err)
	}

	switch application.Valid {
	case database.VALID:
		if err = utility.SendEmail(application.Email, "[城大春祭2023]普通門票確認電郵", func(w io.Writer) error { return regularConfirmationEmailTemplate.Execute(w, application) }); err == nil {
			return nil
		}
		err = fmt.Errorf("failed to send Regular confirmation email for %v: %w", application.Name, err)
	case database.INVALID:
		if err = utility.SendEmail(application.Email, "[城大春祭2023]普通門票付款無效通知電郵", func(w io.Writer) error { return regularNotificationEmailTemplate.Execute(w, application) }); err == nil {
			return nil
		}
		err = fmt.Errorf("failed to send Regular notification email for %v: %w", application.Name, err)
	default:
		err = fmt.Errorf("failed to send Regular notification email for %v: status invalid", application.Name)
	}

	if rollbackErr := application.rollback(database.RECEIVE); rollbackErr != nil {
		return utility.Errors{err, rollbackErr}
	}
	return err

}

func (application *RegularApplication) GenerateTicket() (Ticket, error) {
	if application.Valid == database.INVALID {
		return nil, fmt.Errorf("failed to generate Regular ticket for %v: application invalid", application.Name)
	}

	currentTime := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:   _HARUMATSURI,
		Subject:  _REGULAR_TICKET,
		Audience: jwt.ClaimStrings{application.Name},
		IssuedAt: jwt.NewNumericDate(currentTime),
		ID:       strconv.Itoa(application.Quantity),
	}

	token, err := utility.SignClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to sign claims for %v: %w", application.Name, err)
	}

	ticketInfo := &database.TicketInfo{
		Token:  token,
		Name:   application.Name,
		Key:    utility.SecretKey,
		Status: database.UNUSED,
	}

	if err := ticketInfo.Create(); err != nil {
		return nil, fmt.Errorf("failed to create ticket for %v: %w", application.Name, err)
	}

	qrcode, err := utility.GenerateQRCode(token, color.White, _GREEN)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code for %v, %w", application.Name, err)
	}

	embed := &utility.EmbedFile{
		Name:   fmt.Sprintf("%v.png", utility.HashSha3(qrcode)[:16]),
		Reader: bytes.NewBuffer(qrcode),
	}

	ticket := &TicketInfo{
		TicketInfo: *ticketInfo,
		Embed:      embed,
	}

	regularTicket := &RegularTicket{
		RegularApplication: *application,
		Ticket:             ticket,
	}
	return regularTicket, nil
}

type DiscountApplication struct {
	database.DiscountApplication
}

func (application *DiscountApplication) rollback(status database.ApplicationStatus) error {
	application.Status = status
	if err := application.Update(); err != nil {
		return fmt.Errorf("failed to rollback Discount application status for %v: %w", application.Name, err)
	}
	return nil
}

func (application *DiscountApplication) SendEmail() (err error) {
	application.Status = database.EMAIL_SENDED
	if err = application.Update(); err != nil {
		return fmt.Errorf("failed to update Discount application status for %v: %w", application.Name, err)
	}

	switch application.Valid {
	case database.VALID:
		if err = utility.SendEmail(application.Email, "[城大春祭2023]優惠門票確認電郵", func(w io.Writer) error { return discountConfirmationEmailTemplate.Execute(w, application) }); err == nil {
			return nil
		}
		err = fmt.Errorf("failed to send Discount confirmation email for %v: %w", application.Name, err)
	case database.INVALID:
		if err = utility.SendEmail(application.Email, "[城大春祭2023]優惠門票付款無效通知電郵", func(w io.Writer) error { return discountNotificationEmailTemplate.Execute(w, application) }); err == nil {
			return nil
		}
		err = fmt.Errorf("failed to send Discount notification email for %v: %w", application.Name, err)
	default:
		err = fmt.Errorf("failed to send Discount notification email for %v: status invalid", application.Name)
	}

	if rollbackErr := application.rollback(database.RECEIVE); rollbackErr != nil {
		return utility.Errors{err, rollbackErr}
	}
	return err
}

func (application *DiscountApplication) GenerateTicket() (Ticket, error) {
	if application.Valid == database.INVALID {
		return nil, fmt.Errorf("failed to generate Discount ticket for %v: application invalid", application.Name)
	}

	currentTime := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:   _HARUMATSURI,
		Subject:  _DISCOUNT_TICKET,
		Audience: jwt.ClaimStrings{application.Name},
		IssuedAt: jwt.NewNumericDate(currentTime),
		ID:       strconv.Itoa(application.Quantity),
	}

	token, err := utility.SignClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to sign claims for %v: %w", application.Name, err)
	}

	ticketInfo := &database.TicketInfo{
		Token:  token,
		Name:   application.Name,
		Key:    utility.SecretKey,
		Status: database.UNUSED,
	}

	if err := ticketInfo.Create(); err != nil {
		return nil, fmt.Errorf("failed to create ticket for %v: %w", application.Name, err)
	}

	qrcode, err := utility.GenerateQRCode(token, color.White, _RED)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code for %v, %w", application.Name, err)
	}

	embed := &utility.EmbedFile{
		Name:   fmt.Sprintf("%v.png", utility.HashSha3(qrcode)[:16]),
		Reader: bytes.NewBuffer(qrcode),
	}

	ticket := &TicketInfo{
		TicketInfo: *ticketInfo,
		Embed:      embed,
	}

	discountTicket := &DiscountTicket{
		DiscountApplication: *application,
		Ticket:              ticket,
	}
	return discountTicket, nil
}


func sendVipConfirmationEmail() (sent int, err error) {
	applications_, err := database.VipApplications{}.GetApplicationWithStatus(database.RECEIVE)
	if err != nil {
		return 0, fmt.Errorf("failed to get VIP applications: %w", err)
	}

	applications := make([]*VipApplication, len(applications_))
	for i := range applications {
		applications[i] = &VipApplication{*applications_[i]}
	}

	sent, err = sendConfirmationEmail(applications...)
	if err != nil {
		return sent, fmt.Errorf("failed to send VIP confirmation email: %w", err)
	}
	return sent, nil
}

func sendRegularConfirmationEmail() (sent int, err error) {
	applications_, err := database.RegularApplications{}.GetApplicationWithStatus(database.RECEIVE)
	if err != nil {
		return 0, fmt.Errorf("failed to get Regular application: %w", err)
	}

	applications := make([]*RegularApplication, len(applications_))
	for i := range applications {
		applications[i] = &RegularApplication{*applications_[i]}
	}

	sent, err = sendConfirmationEmail(applications...)
	if err != nil {
		return sent, fmt.Errorf("failed to send Regular confirmation email: %w", err)
	}
	return sent, nil
}

func sendDiscountConfirmationEmail() (sent int, err error) {
	applications_, err := database.DiscountApplications{}.GetApplicationWithStatus(database.RECEIVE)
	if err != nil {
		return 0, fmt.Errorf("failed to get Discount application: %w", err)
	}

	applications := make([]*DiscountApplication, len(applications_))
	for i := range applications_ {
		applications[i] = &DiscountApplication{*applications_[i]}
	}

	sent, err = sendConfirmationEmail(applications...)
	if err != nil {
		return sent, fmt.Errorf("failed to send Regular confirmation email: %w", err)
	}
	return sent, nil
}

func sendConfirmationEmail[T Application](applications ...T) (sent int, err error) {
	errs := utility.Errors{}
	for _, application := range applications {
		err := application.SendEmail()
		if err != nil {
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
