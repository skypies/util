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
func InitGroup(g, csvMembers string) {
	if _,exists := groups[g]; !exists {
		groups[g] = map[string]int{}
	}
	
	for _,e := range strings.Split(csvMembers, ",") {
		groups[g][strings.ToLower(e)] = 1
	}
}

func IsInGroup(group, email string) bool {
	if _,groupExists := groups[group]; !groupExists {
		return false
	}

	_,exists := groups[group][strings.ToLower(email)]
	return exists
}
