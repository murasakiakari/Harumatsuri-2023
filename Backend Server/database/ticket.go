package database

import (
	"backendserver/utility"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
)

type TicketStatus int

const (
	UNUSED TicketStatus = iota
	USED
)

type TicketInfo struct {
	Token  string
	Name   string
	Key    string
	Status TicketStatus
}

func (info *TicketInfo) Create() error {
	statement, err := db.Prepare("insert into Tickets (Token, Name, SecretKey, Status) values (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(info.Token, info.Name, info.Key, info.Status)
	return err
}

func (info *TicketInfo) Read() error {
	row := db.QueryRow("select Name, SecretKey, Status from Tickets where Token=?", info.Token)
	return row.Scan(&info.Name, &info.Key, &info.Status)
}

func (info *TicketInfo) Update() error {
	statement, err := db.Prepare("update Tickets set Status=? where Token=?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(info.Status, info.Token)
	return err
}

func (info *TicketInfo) Delete() error {
	statement, err := db.Prepare("delete from Tickets where Token=?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(info.Token)
	return err
}

func (info *TicketInfo) Validate() (claims *jwt.RegisteredClaims, err error) {
	claims_, err := utility.ValidateTokenWithKey(info.Token, &jwt.RegisteredClaims{}, info.Key)
	if err != nil {
		return nil, err
	}

	claims, ok := claims_.(*jwt.RegisteredClaims)
	if !ok {
		return nil, fmt.Errorf("failed to cast interface")
	}

	return claims, nil
}
