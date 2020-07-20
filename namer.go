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

func (n Namer) RelationshipFKName(relationship schema.Relationship) (name string) {
	return ConvertNonReservedWordToCap(n.NamingStrategy.RelationshipFKName(relationship))
}

func (n Namer) CheckerName(table, column string) (name string) {
	return ConvertNonReservedWordToCap(n.NamingStrategy.CheckerName(table, column))
}

func (n Namer) IndexName(table, column string) (name string) {
	return ConvertNonReservedWordToCap(n.NamingStrategy.IndexName(table, column))
}
