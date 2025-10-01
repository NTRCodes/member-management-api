package members

type Member struct {
	// These are the fields returned from the Members view in the Radon database.
	ID         int     `db:"id"`
	UBCID      string  `db:"ubcid"`
	SSN        *string `db:"ssn" json:"ssn,omitempty"`
	FirstName  *string `db:"first_name" json:"firstName,omitempty"`
	LastName   *string `db:"last_name" json:"lastName,omitempty"`
	Address    *string `db:"address_1" json:"address,omitempty"`
	Address2   *string `db:"address_2" json:"address2,omitempty"`
	City       *string `db:"city" json:"city,omitempty"`
	State      *string `db:"state" json:"state,omitempty"`
	Zip        *string `db:"zip" json:"zip,omitempty"`
	Phone      *string `db:"phone1" json:"phone,omitempty"`
	Phone2     *string `db:"phone2" json:"phone2,omitempty"`
	Email      *string `db:"email" json:"email,omitempty"`
	Local      *string `db:"local" json:"local,omitempty"`
	Status     *string `db:"status" json:"status,omitempty"`
	Class      *string `db:"class" json:"class,omitempty"`
	DOB        *string `db:"dob" json:"dob,omitempty"`
	Gender     *string `db:"gender" json:"gender,omitempty"`
	InitDate   *string `db:"init_date" json:"initDate,omitempty"`
	Delegate   *string `db:"delegate" json:"delegate,omitempty"`
	Language   *string `db:"lang" json:"language,omitempty"`
	Veteran    *string `db:"veteran" json:"veteran,omitempty"`
	LastUpdate *string `db:"stamp" json:"lastUpdate,omitempty"`
}
