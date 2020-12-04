package oracle

import (
	"gorm.io/gorm/schema"
	"strings"
)

type Namer struct {
	schema.NamingStrategy
}

func ConvertNameToFormat(x string) string {
	return strings.ToUpper(x)
}

func (n Namer) TableName(table string) (name string) {
	return ConvertNameToFormat(n.NamingStrategy.TableName(table))
}

func (n Namer) ColumnName(table, column string) (name string) {
	return ConvertNameToFormat(n.NamingStrategy.ColumnName(table, column))
}

func (n Namer) JoinTableName(table string) (name string) {
	return ConvertNameToFormat(n.NamingStrategy.JoinTableName(table))
}

func (n Namer) RelationshipFKName(relationship schema.Relationship) (name string) {
	return ConvertNameToFormat(n.NamingStrategy.RelationshipFKName(relationship))
}

func (n Namer) CheckerName(table, column string) (name string) {
	return ConvertNameToFormat(n.NamingStrategy.CheckerName(table, column))
}

func (n Namer) IndexName(table, column string) (name string) {
	return ConvertNameToFormat(n.NamingStrategy.IndexName(table, column))
}
