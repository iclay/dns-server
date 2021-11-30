package match

import "regexp"

func DomainMatch(domain string, m map[string]interface{}) bool {
	for k, _ := range m {
		if match, _ := regexp.MatchString(k, domain); match {
			return true
		}
	}
	return false
}
