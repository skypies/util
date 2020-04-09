package handlerware

import(
	"strings"
)

const(
	// AdminGroup is the group name for an important group that grants admin privs.
	AdminGroup = "admin"
)

var(
	groups = map[string]map[string]int{}
)

// InitGroup will populate a particular group, given a CSV string of emails.
// The group `admin` is particularly important.
func InitGroup(group, csvMembers string) {
	for _,e := range strings.Split(csvMembers, ",") {
		groups[group][strings.ToLower(e)] = 1
	}
}

func IsInGroup(group, email string) bool {
	_,exists := groups[group][strings.ToLower(email)]
	return exists
}
