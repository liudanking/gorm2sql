package sqlorm

import (
	"backend/cdmake/program"
	"go/parser"
	"go/token"
	"io/ioutil"

	"testing"
)

func TestGenerateCreateTableSql(t *testing.T) {
	fset := token.NewFileSet()
	data, err := ioutil.ReadFile("../testdata/sqlmodel/user.go")
	if err != nil {
		t.Fatal(err)
	}
	f, err := parser.ParseFile(fset, "model.go", string(data), parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	typeSpec, err := program.GetStructByName(f, "UserEmail")
	if err != nil {
		t.Fatal(err)
	}

	ms, err := NewSqlGenerator(typeSpec)
	if err != nil {
		t.Fatal(err)
	}
	sql, err := ms.GetCreateTableSql()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("sql:\n%s", sql)
}
