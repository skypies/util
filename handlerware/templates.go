package handlerware

import(
	"errors"
	"html/template"
	"regexp"
	"time"

	"golang.org/x/net/context"

	"github.com/skypies/util/date"
	"github.com/skypies/util/widget"
)

var(
	// TemplateDir should contain a dir structure of templates. Setting
	// this value is how the user switches on the templates stuff. It
	// must be relative to the appengine module root, which is the git
	// repo root. Setting this to a bad value, or where templates fail
	// to parse, will cause a panic.
	TemplateDir = ""

	Templates *template.Template
)

// Caller *must* call this
func InitTemplates() {
	if TemplateDir != "" {
		Templates = loadTemplates(TemplateDir)
	}
}

// This will PANIC if the dir does not exist or is empty !
func loadTemplates(dir string) *template.Template {
	return widget.ParseRecursive(template.New("").Funcs(templateFuncMap()), dir)
}

// GetFoo: given a context, extracts the object (or panics; should not be optional). May return nil.
func GetTemplates(ctx context.Context) (*template.Template) {
	tmpl, ok := ctx.Value(templatesKey).(*template.Template)
	if (!ok) { panic ("handlerware.GetTemplates: no object found in context") }
	return tmpl
}


// Some standard helper funcs for all our flightish apps
func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"add": templateAdd,
		"km2feet": templateKM2Feet,
		"spacify": templateSpacifyFlightNumber,
		"dict": templateDict,
		"selectdict": templateSelectDict,
		"formatPdt": templateFormatPdt,
	}
}

func templateAdd(a int, b int) int { return a + b }
func templateKM2Feet(x float64) float64 { return x * 3280.84 }
func templateSpacifyFlightNumber(s string) string {
	s2 := regexp.MustCompile("^r:(.+)$").ReplaceAllString(s, "Registration:$1")
	s3 := regexp.MustCompile("^(..)(\\d\\d\\d)$").ReplaceAllString(s2, "$1 $2")
	return regexp.MustCompile("^(..)(\\d\\d)$").ReplaceAllString(s3, "$1  $2")
}
func templateDict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 { return nil, errors.New("invalid dict call")	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i+=2 {
		key, ok := values[i].(string)
		if !ok { return nil, errors.New("dict keys must be strings") }
		dict[key] = values[i+1]
	}
	return dict, nil
}
func templateFormatPdt(t time.Time, format string) string {
	return date.InPdt(t).Format(format)
}

func templateSelectDict(name, dflt string, vals interface{}) map[string]interface{} {
	return map[string]interface{}{
		"Name": name,
		"Default": dflt,
		"Vals": vals,
	}
}
