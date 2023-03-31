package service

import (
	"backendserver/database"
	"fmt"
	"net/http"
	"time"
)

type Response struct {
	Status int
	Body   map[string]string
}

type ResponseT struct {
	Status       int
	ErrorMessage string
}

type ResponseWithBody[T any] struct {
	ResponseT
	Body T
}

type TicketMessages struct {
	Vip      string
	Regular  string
	Discount string
}

type TicketInfoWithStatus struct {
	Type   string
	Name   string
	ID     string
	Status string
}

type Service struct{}

func New() (*Service, error) {
	err := database.Connect()
	if err != nil {
		return nil, err
	}
	return &Service{}, nil
}

func (s *Service) CreateAccount(account database.Account, response *ResponseT) error {
	if err := CreateAccount(&account); err != nil {
		response.Status = http.StatusInternalServerError
		response.ErrorMessage = err.Error()
		return nil
	}

	response.Status = http.StatusOK
	return nil
}

func (s *Service) Login(account database.Account, response *ResponseWithBody[string]) error {
	token, err := Login(&account)
	if err != nil {
		response.Status = http.StatusUnauthorized
		response.ErrorMessage = err.Error()
		return nil
	}

	response.Status = http.StatusOK
	response.Body = token
	return nil
}

func (s *Service) VerifyToken(token string, response *ResponseWithBody[string]) error {
	claims, err := ValidateToken(token)
	if err != nil {
		response.Status = http.StatusUnauthorized
		response.ErrorMessage = err.Error()
		return nil
	}

	response.Status = http.StatusOK
	response.Body = claims.ID
	return nil
}

func (s *Service) CreateTicket(discountApplication database.DiscountApplication, response *ResponseT) error {
	currentTime := time.Now().Unix()
	discountApplication.Name = fmt.Sprintf("%v-%v", discountApplication.Name, currentTime)

	vipApplication := &database.VipApplication{
		ApplicationInfo: database.ApplicationInfo{
			Name:     discountApplication.Name,
			Email:    discountApplication.Email,
			Quantity: discountApplication.Quantity,
			Valid:    discountApplication.Valid,
			Status:   database.RECEIVE,
		},
	}

	if err := vipApplication.Create(); err != nil {
		response.Status = http.StatusInternalServerError
		response.ErrorMessage = "failed to create VIP ticket application: " + err.Error()
		return nil
	}

	regularApplication := &database.RegularApplication{
		ApplicationInfo: database.ApplicationInfo{
			Name:     discountApplication.Name,
			Email:    discountApplication.Email,
			Quantity: discountApplication.Quantity,
			Valid:    discountApplication.Valid,
			Status:   database.RECEIVE,
		},
	}

	if err := regularApplication.Create(); err != nil {
		response.Status = http.StatusInternalServerError
		response.ErrorMessage = "failed to create normal ticket application: " + err.Error()
		return nil
	}

	discountApplication.Status = database.RECEIVE

	if err := discountApplication.Create(); err != nil {
		response.Status = http.StatusInternalServerError
		response.ErrorMessage = "failed to create discount ticket application: " + err.Error()
		return nil
	}

	response.Status = http.StatusOK
	return nil
}

func (s *Service) SendConfirmationEmail(_ *struct{}, response *ResponseWithBody[TicketMessages]) error {
	sent, err := sendVipConfirmationEmail()
	if err != nil {
		response.Body.Vip = fmt.Sprintf("%v VIP confirmation email have send successfully\n\nfollowing error(s) have appear:\n%v", sent, err.Error())
	} else {
		response.Body.Vip = fmt.Sprintf("%v VIP confirmation email have send successfully", sent)
	}

	sent, err = sendRegularConfirmationEmail()
	if err != nil {
		response.Body.Regular = fmt.Sprintf("%v Regular confirmation email have send successfully\n\nfollowing error(s) have appear:\n%v", sent, err.Error())
	} else {
		response.Body.Regular = fmt.Sprintf("%v Regular confirmation email have send successfully", sent)
	}

	sent, err = sendDiscountConfirmationEmail()
	if err != nil {
		response.Body.Discount = fmt.Sprintf("%v Discount confirmation email have send successfully\n\nfollowing error(s) have appear:\n%v", sent, err.Error())
	} else {
		response.Body.Discount = fmt.Sprintf("%v Discount confirmation email have send successfully", sent)
	}

	response.Status = http.StatusOK
	return nil
}

func (s *Service) SendTicket(_ *struct{}, response *ResponseWithBody[TicketMessages]) error {
	sent, err := sendVipTicket()
	if err != nil {
		response.Body.Vip = fmt.Sprintf("%v VIP ticket have send successfully\n\nfollowing error(s) have appear:\n%v", sent, err.Error())
	} else {
		response.Body.Vip = fmt.Sprintf("%v VIP ticket have send successfully", sent)
	}

	sent, err = sendRegularTicket()
	if err != nil {
		response.Body.Regular = fmt.Sprintf("%v Regular ticket have send successfully\n\nfollowing error(s) have appear:\n%v", sent, err.Error())
	} else {
		response.Body.Regular = fmt.Sprintf("%v Regular ticket have send successfully", sent)
	}

	sent, err = sendDiscountTicket()
	if err != nil {
		response.Body.Discount = fmt.Sprintf("%v Discount ticket have send successfully\n\nfollowing error(s) have appear:\n%v", sent, err.Error())
	} else {
		response.Body.Discount = fmt.Sprintf("%v Discount ticket have send successfully", sent)
	}

	response.Status = http.StatusOK
	return nil
}

// General ticket

func (s *Service) GetTicketInfo(token string, response *ResponseWithBody[TicketInfoWithStatus]) error {
	claims, status, err := getTicketInfo(token)
	if err != nil {
		response.Status = http.StatusBadRequest
		response.ErrorMessage = "failed to get ticket info: " + err.Error()
		return nil
	}

	if len(claims.Audience) != 1 {
		response.Status = http.StatusInternalServerError
		response.ErrorMessage = "failed to get ticket info: the length of field audience is not 1"
	}

	response.Status = http.StatusOK
	response.Body.Type = claims.Subject
	response.Body.Name = claims.Audience[0]
	response.Body.ID = claims.ID
	response.Body.Status = status
	return nil
}

func (s *Service) UseTicket(token string, response *ResponseWithBody[string]) error {
	claims, err := useTicket(token)
	if err != nil {
		response.Status = http.StatusInternalServerError
		response.ErrorMessage = "failed to update ticket status: " + err.Error()
		return nil
	}

	response.Status = http.StatusOK
	response.Body = claims.ID
	return nil
}

func (s *Service) StartQueuing(peopleAdmitted int, response *ResponseWithBody[[]byte]) error {
	message, err := queuingSystem.StartQueuing(peopleAdmitted)
	if err != nil {
		response.Status = http.StatusInternalServerError
		response.ErrorMessage = "failed to start queuing: " + err.Error()
		return nil
	}

	response.Status = http.StatusOK
	response.Body = message
	return nil
}

func (s *Service) AllowEntry(peopleAdmitted int, response *ResponseWithBody[[]byte]) error {
	message, err := queuingSystem.AllowEntry(peopleAdmitted)
	if err != nil {
		response.Status = http.StatusInternalServerError
		response.ErrorMessage = "failed to update entry allowance: " + err.Error()
		return nil
	}

	response.Status = http.StatusOK
	response.Body = message
	return nil
}

func (s *Service) SetOffset(offset float64, response *ResponseWithBody[[]byte]) error {
	message, err := queuingSystem.SetOffset(offset)
	if err != nil {
		response.Status = http.StatusInternalServerError
		response.ErrorMessage = "failed to set offset: " + err.Error()
		return nil
	}

	response.Status = http.StatusOK
	response.Body = message
	return nil
}

func (s *Service) UpdateFirstEntry(peopleEntries int, response *ResponseWithBody[[]byte]) error {
	message, err := queuingSystem.UpdateFirstEntry(peopleEntries)
	if err != nil {
		response.Status = http.StatusInternalServerError
		response.ErrorMessage = "failed to update first entry: " + err.Error()
		return nil
	}

	response.Status = http.StatusOK
	response.Body = message
	return nil
}

func (s *Service) UpdateSecondEntry(peopleEntries int, response *ResponseWithBody[[]byte]) error {
	message, err := queuingSystem.UpdateSecondEntry(peopleEntries)
	if err != nil {
		response.Status = http.StatusInternalServerError
		response.ErrorMessage = "failed to update second entry: " + err.Error()
		return nil
	}

	response.Status = http.StatusOK
	response.Body = message
	return nil
}

func (s *Service) GetOrder(number string, response *ResponseWithBody[int]) error {
	order := queuingSystem.GetOrder(number)
	if order == -1 {
		response.Status = http.StatusBadRequest
		return nil
	}

	response.Status = http.StatusOK
	response.Body = order
	return nil
}

func (s *Service) GetMessage(_ *struct{}, response *ResponseWithBody[[]byte]) error {
	message, err := queuingSystem.GetMessage()
	if err != nil {
		response.Status = http.StatusInternalServerError
		response.ErrorMessage = "failed to get message: " + err.Error()
		return nil
	}

	response.Status = http.StatusOK
	response.Body = message
	return nil
}
