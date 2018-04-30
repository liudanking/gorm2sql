//go:generate go-bindata -o template_data.go -pkg gencode ../template/
package gencode

import (
	"bytes"
	"text/template"
)

func RenderTemplate(tpl string, data interface{}, gofmt bool) (string, error) {
	t, err := template.New("").Parse(tpl)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err = t.Execute(buf, data); err != nil {
		return "", err
	}

	// if gofmt {
	// 	src, err := format.Source(buf.Bytes())
	// 	return string(src), err
	// }

	return buf.String(), nil
}
