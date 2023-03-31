package database

// abstract of the VipApplications table
type VipApplications struct {}

func (applications VipApplications) GetApplicationWithStatus(status ApplicationStatus) ([]*VipApplication, error) {
	rows, err := db.Query("select Name, Email, Quantity, Valid, Status from VipApplications where Status=?", status)
	if err != nil {
		return nil, err
	}

	vipApplications := make([]*VipApplication, 0, 16)
	for rows.Next() {
		application := &VipApplication{}
		if err = rows.Scan(&application.Name, &application.Email, &application.Quantity, &application.Valid, &application.Status); err != nil {
			return nil, err
		}
		vipApplications = append(vipApplications, application)
	}

	return vipApplications, nil
}

// abstract of the RegularApplications table
type RegularApplications struct {}

func (applications RegularApplications) GetApplicationWithStatus(status ApplicationStatus) ([]*RegularApplication, error) {
	rows, err := db.Query("select Name, Email, Quantity, Valid, Status from RegularApplications where Status=?", status)
	if err != nil {
		return nil, err
	}

	regularApplications := make([]*RegularApplication, 0, 16)
	for rows.Next() {
		application := &RegularApplication{}
		if err = rows.Scan(&application.Name, &application.Email, &application.Quantity, &application.Valid, &application.Status); err != nil {
			return nil, err
		}
		regularApplications = append(regularApplications, application)
	}
	return regularApplications, nil
}

type DiscountApplications struct {}

func (applications DiscountApplications) GetApplicationWithStatus(status ApplicationStatus) ([]*DiscountApplication, error) {
	rows, err := db.Query("select Name, Email, Quantity, SID, Valid, Status from DiscountApplications where Status=?", status)
	if err != nil {
		return nil, err
	}

	discountApplications := make([]*DiscountApplication, 0, 16)
	for rows.Next() {
		application := &DiscountApplication{}
		if err = rows.Scan(&application.Name, &application.Email, &application.Quantity, &application.SID, &application.Valid, &application.Status); err != nil {
			return nil, err
		}
		discountApplications = append(discountApplications, application)
	}
	return discountApplications, nil
}
