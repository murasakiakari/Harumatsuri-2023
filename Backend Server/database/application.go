package database

type Valid int

const (
	INVALID Valid = iota
	VALID
)

type ApplicationStatus int

const (
	RECEIVE ApplicationStatus = iota
	EMAIL_SENDED
	TICKET_SENDED
)

type ApplicationInfo struct {
	Name     string
	Email    string
	Quantity int
	Valid    Valid
	Status   ApplicationStatus
}

type VipApplication struct {
	ApplicationInfo
}

func (application *VipApplication) Create() error {
	statement, err := db.Prepare("insert into VipApplications (Name, Email, Quantity, Valid, Status) values (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(application.Name, application.Email, application.Quantity, application.Valid, application.Status)
	return err
}

func (application *VipApplication) Update() error {
	statement, err := db.Prepare("update VipApplications set Status=? where Name=?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(application.Status, application.Name)
	return err
}

type RegularApplication struct {
	ApplicationInfo
}

func (application *RegularApplication) Create() error {
	statement, err := db.Prepare("insert into RegularApplications (Name, Email, Quantity, Valid, Status) values (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(application.Name, application.Email, application.Quantity, application.Valid, application.Status)
	return err
}

func (application *RegularApplication) Update() error {
	statement, err := db.Prepare("update RegularApplications set Status=? where Name=?")
	if err != nil {
		return err
	}

	_, err = statement.Exec(application.Status, application.Name)
	return err
}

type DiscountApplication struct {
	ApplicationInfo
	SID string
}

func (application *DiscountApplication) Create() error {
	statement, err := db.Prepare("insert into DiscountApplications (Name, Email, Quantity, SID, Valid, Status) values (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(application.Name, application.Email, application.Quantity, application.SID, application.Valid, application.Status)
	return err
}

func (application *DiscountApplication) Update() error {
	statement, err := db.Prepare("update DiscountApplications set Status=? where Name=?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(application.Status, application.Name)
	return err
}
