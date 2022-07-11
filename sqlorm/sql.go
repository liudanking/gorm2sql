package sqlorm

import (
	"errors"
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"github.com/liudanking/gorm2sql/util"

	"github.com/pinzolo/casee"

	log "github.com/liudanking/goutil/logutil"
)

type SqlGenerator struct {
	structName      string
	modelType       *ast.StructType
	customTableName string
}

func NewSqlGenerator(typeSpec *ast.TypeSpec, tableName string) (*SqlGenerator, error) {
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return nil, errors.New("typeSpec is not struct type")
	}

	return &SqlGenerator{
		structName:      typeSpec.Name.Name,
		modelType:       structType,
		customTableName: tableName,
	}, nil
}

func (ms *SqlGenerator) GetCreateTableSql() (string, error) {
	var tags []string
	var primaryKeys []string
	indices := map[string][]string{}
	uniqIndces := map[string][]string{}
	for _, field := range ms.getStructFieds(ms.modelType) {
		switch t := field.Type.(type) {
		case *ast.Ident:
			tag, err := generateSqlTag(field)
			if err != nil {
				log.Warning("generateSqlTag [%s] failed:%v", t.Name, err)
			} else {
				tags = append(tags, fmt.Sprintf("%s %s", getColumnName(field), tag))
			}
		case *ast.SelectorExpr:
			tag, err := generateSqlTag(field)
			if err != nil {
				log.Warning("generateSqlTag [%s] failed:%v", t.Sel.Name, err)
			} else {
				tags = append(tags, fmt.Sprintf("%s %s", getColumnName(field), tag))
			}
		default:
			log.Warning("field %s not supported, ignore", util.GetFieldName(field))
		}

		columnName := getColumnName(field)
		if isPrimaryKey(field) {
			primaryKeys = append(primaryKeys, columnName)
		}

		sqlSettings := ParseTagSetting(util.GetFieldTag(field, "sql").Name)
		if idxName, ok := sqlSettings["INDEX"]; ok {
			keys := indices[idxName]
			keys = append(keys, columnName)
			indices[idxName] = keys
		}
		if idxName, ok := sqlSettings["UNIQUE_INDEX"]; ok {
			keys := uniqIndces[idxName]
			keys = append(keys, columnName)
			uniqIndces[idxName] = keys
		}

	}

	var primaryKeyStr string
	if len(primaryKeys) > 0 {
		primaryKeyStr = fmt.Sprintf("PRIMARY KEY (%v)", strings.Join(primaryKeys, ", "))
	}

	indicesStrs := []string{}
	for idxName, keys := range indices {
		indicesStrs = append(indicesStrs, fmt.Sprintf("INDEX %s (%s)", idxName, strings.Join(keys, ", ")))
	}

	uniqIndicesStrs := []string{}
	for idxName, keys := range uniqIndces {
		uniqIndicesStrs = append(uniqIndicesStrs, fmt.Sprintf("UNIQUE INDEX %s (%s)", idxName, strings.Join(keys, ", ")))
	}

	options := []string{
		"engine=innodb",
		"DEFAULT charset=utf8mb4",
	}

	return fmt.Sprintf(`CREATE TABLE %v 
(
  %v,
  %v
) %v;`,
		"`"+ms.tableName()+"`",
		strings.Join(append(tags, append(indicesStrs, uniqIndicesStrs...)...), ",\n  "),
		primaryKeyStr,
		strings.Join(options, " ")), nil
}

func (ms *SqlGenerator) getStructFieds(node ast.Node) []*ast.Field {
	var fields []*ast.Field
	nodeType, ok := node.(*ast.StructType)
	if !ok {
		return nil
	}
	for _, field := range nodeType.Fields.List {
		if util.GetFieldTag(field, "sql").Name == "-" {
			continue
		}

		switch t := field.Type.(type) {
		case *ast.Ident:
			if t.Obj != nil && t.Obj.Kind == ast.Typ {
				if typeSpec, ok := t.Obj.Decl.(*ast.TypeSpec); ok {
					fields = append(fields, ms.getStructFieds(typeSpec.Type)...)
				}
			} else {
				fields = append(fields, field)
			}
		case *ast.SelectorExpr:
			fields = append(fields, field)
		default:
			log.Warning("filed %s not supported, ignore", util.GetFieldName(field))
		}
	}

	return fields
}

func (ms *SqlGenerator) tableName() string {
	if len(ms.customTableName) > 0 {
		return ms.customTableName
	}
	return casee.ToSnakeCase(ms.structName)
}

func generateSqlTag(field *ast.Field) (string, error) {
	var sqlType string
	var err error

	tagStr := util.GetFieldTag(field, "sql").Name
	sqlSettings := ParseTagSetting(tagStr)

	if value, ok := sqlSettings["TYPE"]; ok {
		sqlType = value
	}

	if _, found := sqlSettings["NOT NULL"]; !found { // default: not null
		sqlSettings["NOT NULL"] = "NOT NULL"
	}

	additionalType := sqlSettings["NOT NULL"] + " " + sqlSettings["UNIQUE"]
	if value, ok := sqlSettings["DEFAULT"]; ok {
		additionalType = additionalType + " DEFAULT " + value
	}

	if sqlType == "" {
		var size = 128

		if value, ok := sqlSettings["SIZE"]; ok {
			size, _ = strconv.Atoi(value)
		}

		_, autoIncrease := sqlSettings["AUTO_INCREMENT"]
		if isPrimaryKey(field) {
			autoIncrease = true
		}

		decimalStr := ""
		if value, ok := sqlSettings["TYPE"]; ok {
			if strings.Contains(value, "decimal(") {
				decimalStr = value
			}
		}

		sqlType, err = mysqlTag(field, size, autoIncrease, decimalStr)
		if err != nil {
			log.Warning("get mysql field tag failed:%v", err)
			return "", err
		}
	}

	if strings.TrimSpace(additionalType) == "" {
		return sqlType, nil
	} else {
		return fmt.Sprintf("%v %v", sqlType, additionalType), nil
	}

}

func getColumnName(field *ast.Field) string {
	tagStr := util.GetFieldTag(field, "gorm").Name
	gormSettings := ParseTagSetting(tagStr)
	if columnName, ok := gormSettings["COLUMN"]; ok {
		return columnName
	}

	if len(field.Names) > 0 {
		return fmt.Sprintf("`%s`", casee.ToSnakeCase(field.Names[0].Name))
	}

	return ""
}

func isPrimaryKey(field *ast.Field) bool {
	tagStr := util.GetFieldTag(field, "gorm").Name
	gormSettings := ParseTagSetting(tagStr)
	if _, ok := gormSettings["PRIMARY_KEY"]; ok {
		return true
	}

	if len(field.Names) > 0 && strings.ToUpper(field.Names[0].Name) == "ID" {
		return true
	}

	return false
}

func mysqlTag(field *ast.Field, size int, autoIncrease bool, decimalType string) (string, error) {
	typeName := ""
	switch t := field.Type.(type) {
	case *ast.Ident:
		typeName = t.Name
	case *ast.SelectorExpr:
		typeName = t.Sel.Name
	default:
		return "", errors.New(fmt.Sprintf("field %s not supported", util.GetFieldName(field)))
	}

	switch typeName {
	case "bool":
		return "boolean", nil
	case "int", "int8", "int16", "int32", "uint", "uint8", "uint16", "uint32", "uintptr":
		if autoIncrease {
			return "int AUTO_INCREMENT", nil
		}
		return "int", nil
	case "int64", "uint64":
		if autoIncrease {
			return "bigint AUTO_INCREMENT", nil
		}
		return "bigint", nil
	case "float32", "float64":
		return "double", nil
	case "string", "NullString":
		if size > 0 && size < 65532 {
			return fmt.Sprintf("varchar(%d)", size), nil
		}
		return "longtext", nil
	case "Time":
		return "datetime", nil
	case "Decimal":
		if len(decimalType) > 0 {
			return decimalType, nil
		}
		return "decimal(8,2)", nil
	default:
		return "", errors.New(fmt.Sprintf("type %s not supported", typeName))

	}
}

func ParseTagSetting(str string) map[string]string {
	tags := strings.Split(str, ";")
	setting := map[string]string{}
	for _, value := range tags {
		v := strings.Split(value, ":")
		k := strings.TrimSpace(strings.ToUpper(v[0]))
		if len(v) == 2 {
			setting[k] = v[1]
		} else {
			setting[k] = k
		}
	}
	return setting
}
