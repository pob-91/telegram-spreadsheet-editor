package utils

import "regexp"

func IsFinancial(str string) bool {
	re := regexp.MustCompile(`^£?\d+(\.\d+)?$`)
	return re.MatchString(str)
}
