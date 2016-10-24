package date

import(
	"fmt"
	"net/http"
	"strconv"
)

func formValueInt64EatErrs(r *http.Request, name string) int64 {
	val,_ := strconv.ParseInt(r.FormValue(name), 10, 64)
	return val
}

// If fields absent or blank, returns {0.0, 0.0}
func FormValueTimeOfDayRange(r *http.Request, stem string) (TimeOfDayRange,error) {
	return ParseTimeOfDay(r.FormValue(stem+"_start"), r.FormValue(stem+"_len"),)
}

func (tr TimeOfDayRange)ToCGIArgs(stem string) string {
	return fmt.Sprintf("%s_start=%02d:%02d&%s_len=%s", stem, tr.Hour, tr.Minute, stem, tr.Length)
}
