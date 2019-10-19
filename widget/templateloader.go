package widget

import(
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

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
