package gencode

import "testing"

func TestAbsPath(t *testing.T) {
	path := "~/code/golang/gopath/src/backend/genginservice"
	newPath, err := AbsPath(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", newPath)

	path = "./code/golang/gopath/src/backend/genginservice"
	newPath, err = AbsPath(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", newPath)
}
