package db

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const structTpl = `// 自动生成，请勿修改
package {{.Package}}

type {{.TableName | ToUpperCamelCase}} struct { {{range .Columns}}{{$length := len .Comment}}{{if gt $length 0}}
	// {{.Comment}}{{end}}
	{{.Name}} {{.DataType}} {{.Tag}}{{end}}
}

func (t {{.TableName | ToUpperCamelCase}}) TableName() string {
	return "{{.TableName | ToLower}}"
}
`

type StructColumn struct {
	Name     string
	DataType string
	Comment  string
	Tag      string
}

type StructTemplateDB struct {
	TableName string
	Package   string
	Columns   []*StructColumn
}

func UnderscoreToUpperCamelCase(word string) string {
	word = strings.ToLower(strings.ReplaceAll(word, "_", " "))
	word = strings.ReplaceAll(strings.Title(word), " ", "")

	return word
}

func ToLower(word string) string {
	return strings.ToLower(word)
}

func GetTemplate() *template.Template {
	tpl := template.Must(template.New("sql2gostruct").Funcs(template.FuncMap{
		"ToLower":          ToLower,
		"ToUpperCamelCase": UnderscoreToUpperCamelCase,
	}).Parse(structTpl))

	return tpl
}

func ReGenDir(output string) error {
	output = filepath.Base(output)
	dir, err := os.Stat(output)
	if err != nil {
		return err
	}

	if dir.IsDir() {
		_ = os.RemoveAll(output)

		err = os.MkdirAll(output, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}
