package handlerware

import(
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/skypies/util/date"
)

var(
	Templates *template.Template
)

// InitTemplates *must* be invoked by the caller. The single arg
// should contain a dir structure of template. It must be relative to
// the appengine module root, which is the git repo root. Setting this
// to a bad value, or where templates fail to parse, will cause a
// panic.
func InitTemplates(templateDir string) {
	Templates = loadTemplates(templateDir)
}

// This will PANIC if the dir does not exist or is empty !
func loadTemplates(dir string) *template.Template {
	return ParseRecursive(template.New("").Funcs(templateFuncMap()), dir)
}

// GetFoo: given a context, extracts the object (or panics; should not be optional). May return nil.
func GetTemplates(ctx context.Context) (*template.Template) {
	tmpl, ok := ctx.Value(templatesKey).(*template.Template)
	if (!ok) { panic ("handlerware.GetTemplates: no object found in context") }
	return tmpl
}

// ParseRecursive walks the directory structure, loading all the files
// it finds. Will panic on failure. Follows symlinks.
func ParseRecursive(t *template.Template, dir string) *template.Template {
	dirStack := []string{dir}
	fileStack := []string{}

	for {
		if len(dirStack) == 0 { break }

		dir := dirStack[0]
		dirStack = dirStack[1:]

		contents,err := ioutil.ReadDir(dir)
		if err != nil {
			str := ""
			filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
				if err != nil { return err }
				if info.IsDir() {
					str += info.Name() + "/\n"
				}
				return nil
			})

			panic(fmt.Errorf("ReadDir: %v\n\nContents of .:\n--\n%s", err, str))
		}

		for _,f := range contents {
			if strings.HasPrefix(f.Name(), ".") {
				continue
			}
			if f.IsDir() || (f.Mode() & os.ModeSymlink != 0) {
				dirStack = append(dirStack, dir+"/"+f.Name())
			} else {
				fileStack = append(fileStack, dir+"/"+f.Name())
			}
		}
	}

	return template.Must(t.ParseFiles(fileStack...))
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
