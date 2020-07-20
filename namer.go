package oracle

import (
	"strings"

	"gorm.io/gorm/schema"
)

type Namer struct {
	schema.NamingStrategy
}

func ConvertNonReservedWordToCap(x string) string {
	if !IsReservedWord(x) {
		x = strings.ToUpper(x)
	}
	return x
}

func (n Namer) TableName(table string) (name string) {
	return ConvertNonReservedWordToCap(n.NamingStrategy.TableName(table))
}

func (n Namer) JoinTableName(table string) (name string) {
	return ConvertNonReservedWordToCap(n.NamingStrategy.JoinTableName(table))
}
