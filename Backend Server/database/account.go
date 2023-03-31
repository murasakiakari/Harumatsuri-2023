package database

type Account struct {
	Username   string
	Password   string
	Permission string
}

func (account *Account) Create() error {
	statement, err := db.Prepare("insert into Accounts (Username, Password, Permission) values (?, ?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(account.Username, account.Password, account.Permission)
	return err
}

func (account *Account) Read() error {
	row := db.QueryRow("select Password, Permission from Accounts where Username=?", account.Username)
	return row.Scan(&account.Password, &account.Permission)
}
