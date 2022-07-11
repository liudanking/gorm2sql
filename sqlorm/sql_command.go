package sqlorm

import (
	"errors"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/liudanking/gorm2sql/gencode"
	"github.com/liudanking/gorm2sql/program"

	log "github.com/liudanking/goutil/logutil"
	"github.com/urfave/cli"
)

func SqlCommand() cli.Command {
	return cli.Command{
		Name:  "sql",
		Usage: "generate sql from golang model struct",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "file, f",
				Usage: "source file or dir, default: current dir",
			},
			cli.StringFlag{
				Name:  "struct, s",
				Usage: "struct name or pattern: https://golang.org/pkg/path/filepath/#Match",
			},
			cli.StringFlag{
				Name:  "out, o",
				Usage: "output file",
			},
			cli.StringFlag{
				Name:  "table_name, t",
				Usage: "custom table name",
			},
		},
		Action: SqlCommandAction,
	}
}

func SqlCommandAction(c *cli.Context) error {
	file := c.String("file")
	if file == "" {
		file, _ = os.Getwd()
	}
	fi, err := os.Stat(file)
	if err != nil {
		log.Warning("get file info [%s] failed:%v", file, err)
		return err
	}

	pattern := c.String("struct")
	if pattern == "" {
		return errors.New("struct is empty")
	}

	out := c.String("out")
	if out == "" {
		return errors.New("output file is empty")
	}

	matchFunc := func(structName string) bool {
		match, _ := filepath.Match(pattern, structName)
		return match
	}

	tableName := c.String("table_name")

	var types []*ast.TypeSpec
	if !fi.IsDir() {
		fset := token.NewFileSet()
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Warning("read [file:%s] failed:%v", file, err)
			return err
		}
		f, err := parser.ParseFile(fset, file, string(data), parser.ParseComments)
		if err != nil {
			log.Warning("parse [file:%s] failed:%v", file, err)
			return err
		}
		types = program.FindMatchStruct([]*ast.File{f}, matchFunc)
	} else {
		absPath, err := gencode.AbsPath(file)
		if err != nil {
			log.Warning("get [path:%s] absPath failed:%v", file, err)
			return err
		}
		srcPkg, err := build.ImportDir(absPath, build.IgnoreVendor)
		if err != nil {
			log.Warning("get package [%s] info failed:%v", absPath, err)
			return err
		}

		prog, err := program.NewProgram([]string{srcPkg.ImportPath})
		if err != nil {
			log.Warning("new program failed:%v", err)
			return err
		}
		pi, err := prog.GetPkgByName(srcPkg.ImportPath)
		if err != nil {
			log.Warning("get package [%s] failed:%v", srcPkg.ImportPath, err)
			return err
		}
		types = program.FindMatchStruct(pi.Files, matchFunc)
	}

	log.Info("get %d matched struct", len(types))

	sqls := []string{}
	for _, typ := range types {
		ms, err := NewSqlGenerator(typ, tableName)
		if err != nil {
			log.Warning("create model struct failed:%v", err)
			return err
		}

		sql, err := ms.GetCreateTableSql()
		if err != nil {
			log.Warning("generate sql failed:%v", err)
			return err
		}

		sqls = append(sqls, sql)
	}

	return ioutil.WriteFile(out, []byte(strings.Join(sqls, "\n\n")), 0666)
}
