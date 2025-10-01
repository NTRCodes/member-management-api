package privacy

import "strconv"

// LastFourSSN should take a full SSN and return a masked version
// Example: "123456789" -> "XXX-XX-6789"
func LastFourSSN(ssn string) string {
	if len(ssn) >= 4 {
		return ssn[len(ssn)-4:]
	}
	return ""
}

// IsValidSSN should validate SSN format (9 digits)
func IsValidSSN(ssn string) bool {
	_, err := strconv.Atoi(ssn)
	return len(ssn) == 9 && err == nil
}
