package chaincode

import (
	"strings"
	"time"
)

type VaccinationDate time.Time

func (vd *VaccinationDate) MarshalJSON() ([]byte, error) {
	str := time.Time(*vd).Format("2006-01-02 15")
	return []byte("\"" + str + "\""), nil
}

func (vd *VaccinationDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse("2006-01-02 15", s)
	*vd = VaccinationDate(nt)
	return
}
