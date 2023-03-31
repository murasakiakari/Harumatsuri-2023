package backend

import (
	"fmt"
	"gatewayserver/utility"
	"net/http"
	"net/rpc"
)

type Response struct {
	Status int
	Body   map[string]string
}

type ResponseT struct {
	Status       int
	ErrorMessage string
}

func (response *ResponseT) Unwrap() error {
	if response.Status != http.StatusOK {
		return fmt.Errorf("backend response status: %v with error: %v", http.StatusText(response.Status), response.ErrorMessage)
	}
	return nil
}

type ResponseWithBody[T any] struct {
	ResponseT
	Body T
}

func (response *ResponseWithBody[T]) Unwrap() (body T, err error) {
	if response.Status != http.StatusOK {
		return body, fmt.Errorf("backend response status: %v with error: %v", http.StatusText(response.Status), response.ErrorMessage)
	}
	return response.Body, nil
}

type TicketMessage struct {
	Vip      string `json:"vip"`
	Regular  string `json:"regular"`
	Discount string `json:"discount"`
}

type TicketInfoWithStatus struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	ID     string `json:"id"`
	Status string `json:"status"`
}

type Account struct {
	Username   string `form:"username" binding:"required"`
	Password   string `form:"password" binding:"required"`
	Permission string
}

type TicketApplication struct {
	Name     string `form:"name" binding:"required"`
	Email    string `form:"email" binding:"required"`
	SID      string `form:"sid" binding:"required"`
	Valid    int    `form:"valid"`
	Quantity int    `form:"quantity" binding:"required"`
}

func connect() (*rpc.Client, error) {
	return rpc.Dial("tcp", utility.Config.Backend.Address)
}

func CreateAccount(account *Account) error {
	client, err := connect()
	if err != nil {
		return err
	}
	defer client.Close()

	response := &ResponseT{}
	if err = client.Call("Service.CreateAccount", account, response); err != nil {
		return err
	}
	return response.Unwrap()
}

func Login(account *Account) (token string, err error) {
	client, err := connect()
	if err != nil {
		return "", err
	}
	defer client.Close()

	response := &ResponseWithBody[string]{}
	if err = client.Call("Service.Login", account, response); err != nil {
		return "", err
	}
	return response.Unwrap()
}

func VerifyToken(token string) (permission string, err error) {
	client, err := connect()
	if err != nil {
		return "", err
	}
	defer client.Close()

	response := &ResponseWithBody[string]{}
	if err = client.Call("Service.VerifyToken", token, response); err != nil {
		return "", err
	}
	return response.Unwrap()
}

func CreateTicket(application *TicketApplication) error {
	client, err := connect()
	if err != nil {
		return err
	}
	defer client.Close()

	response := &ResponseT{}
	if err = client.Call("Service.CreateTicket", application, response); err != nil {
		return err
	}
	return response.Unwrap()
}

func SendConfirmationEmail() (ticketMessage *TicketMessage, err error) {
	client, err := connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	response := &ResponseWithBody[*TicketMessage]{Body: &TicketMessage{}}
	if err = client.Call("Service.SendConfirmationEmail", &struct{}{}, response); err != nil {
		return nil, err
	}
	return response.Unwrap()
}

func SendTicket() (ticketMessage *TicketMessage, err error) {
	client, err := connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	response := &ResponseWithBody[*TicketMessage]{Body: &TicketMessage{}}
	if err = client.Call("Service.SendTicket", &struct{}{}, response); err != nil {
		return nil, err
	}
	return response.Unwrap()
}

func GetTicketInfo(token string) (ticketInfo *TicketInfoWithStatus, err error) {
	client, err := connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	response := &ResponseWithBody[*TicketInfoWithStatus]{Body: &TicketInfoWithStatus{}}
	if err = client.Call("Service.GetTicketInfo", token, response); err != nil {
		return nil, err
	}
	return response.Unwrap()
}

func UseTicket(token string) (nTicketUsed string, err error) {
	client, err := connect()
	if err != nil {
		return "", err
	}
	defer client.Close()

	response := &ResponseWithBody[string]{}
	if err = client.Call("Service.UseTicket", token, response); err != nil {
		return "", err
	}
	return response.Unwrap()
}

func StartQueuing(peopleAdmitted int) (message []byte, err error) {
	client, err := connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	response := &ResponseWithBody[[]byte]{}
	if err = client.Call("Service.StartQueuing", peopleAdmitted, response); err != nil {
		return nil, err
	}
	return response.Unwrap()
}

func AllowEntry(peopleAdmitted int) (message []byte, err error) {
	client, err := connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	response := &ResponseWithBody[[]byte]{}
	if err = client.Call("Service.AllowEntry", peopleAdmitted, response); err != nil {
		return nil, err
	}
	return response.Unwrap()
}

func SetOffset(offset float64) (message []byte, err error) {
	client, err := connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	response := &ResponseWithBody[[]byte]{}
	if err = client.Call("Service.SetOffset", offset, response); err != nil {
		return nil, err
	}
	return response.Unwrap()
}

func UpdateFirstEntry(peopleEntries int) (message []byte, err error) {
	client, err := connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	response := &ResponseWithBody[[]byte]{}
	if err = client.Call("Service.UpdateFirstEntry", peopleEntries, response); err != nil {
		return nil, err
	}
	return response.Unwrap()
}

func UpdateSecondEntry(peopleEntries int) (message []byte, err error) {
	client, err := connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	response := &ResponseWithBody[[]byte]{}
	if err = client.Call("Service.UpdateSecondEntry", peopleEntries, response); err != nil {
		return nil, err
	}
	return response.Unwrap()
}

func GetOrder(number string) (order int, err error) {
	client, err := connect()
	if err != nil {
		return -1, err
	}
	defer client.Close()

	response := &ResponseWithBody[int]{}
	if err = client.Call("Service.GetOrder", number, response); err != nil {
		return -1, err
	}
	return response.Unwrap()
}

func GetMessage() (message []byte, err error) {
	client, err := connect()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	response := &ResponseWithBody[[]byte]{}
	if err = client.Call("Service.GetMessage", &struct{}{}, response); err != nil {
		return nil, err
	}
	return response.Unwrap()
}
