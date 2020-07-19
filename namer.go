package oracle

import (
	"strings"

	"gorm.io/gorm/schema"
)

type Namer struct {
	schema.NamingStrategy
}

func (n Namer) TableName(table string) (name string) {
	name = n.NamingStrategy.TableName(table)
	if !IsReservedWord(name) {
		name = strings.ToUpper(name)
	}
	return
}

func (n Namer) ColumnName(table, column string) (name string) {
	name = n.NamingStrategy.ColumnName(table, column)
	if !IsReservedWord(name) {
		name = strings.ToUpper(name)
	}
	return
}

func (n Namer) JoinTableName(table string) (name string) {
	name = n.NamingStrategy.JoinTableName(table)
	if !IsReservedWord(name) {
		name = strings.ToUpper(name)
	}
	return
}

func (n Namer) RelationshipFKName(relationship schema.Relationship) (name string) {
	name = n.NamingStrategy.RelationshipFKName(relationship)
	name = strings.ToUpper(name)
	return
}

func (n Namer) CheckerName(table, column string) (name string) {
	name = n.NamingStrategy.CheckerName(table, column)
	name = strings.ToUpper(name)
	return
}

func (n Namer) IndexName(table, column string) (name string) {
	name = n.NamingStrategy.IndexName(table, column)
	name = strings.ToUpper(name)
	return
}
