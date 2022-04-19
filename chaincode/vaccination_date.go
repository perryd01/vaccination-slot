package chaincode

import (
	"strings"
	"time"
)

// VaccinationDate is a simplified date format to identify specific occasions
type VaccinationDate time.Time

// MarshalJSON marshals the date into 2006-01-02 format
func (vd *VaccinationDate) MarshalJSON() ([]byte, error) {
	str := time.Time(*vd).Format("2006-01-02")
	return []byte("\"" + str + "\""), nil
}

// UnmarshalJSON unmarshals date from 2006-01-02 format
func (vd *VaccinationDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse("2006-01-02", s)
	*vd = VaccinationDate(nt)
	return
}
