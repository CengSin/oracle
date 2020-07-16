package oracle

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
)

type Migrator struct {
	migrator.Migrator
}

func (m Migrator) CurrentDatabase() (name string) {
	m.DB.Raw(`SELECT ORA_DATABASE_NAME as "Current Database" FROM DUAL`).Row().Scan(&name)
	return
}

func (m Migrator) CreateTable(values ...interface{}) error {
	m.TryRemoveOnUpdate(values)
	return m.Migrator.CreateTable(values...)
}

func (m Migrator) DropTable(values ...interface{}) error {
	values = m.ReorderModels(values, false)
	tx := m.DB.Session(&gorm.Session{})
	for i := len(values) - 1; i >= 0; i-- {
		if err := m.RunWithValue(values[i], func(stmt *gorm.Statement) error {
			return tx.Exec("DROP TABLE ? CASCADE CONSTRAINTS", clause.Table{Name: stmt.Table}).Error
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m Migrator) HasTable(value interface{}) bool {
	var count int64

	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		return m.DB.Raw("SELECT count(*) FROM user_tables WHERE table_name = UPPER(?)", stmt.Table).Row().Scan(&count)
	})

	return count > 0
}

func (m Migrator) RenameTable(oldName, newName interface{}) error {
	var oldTable, newTable string
	if v, ok := oldName.(string); ok {
		oldTable = v
	} else {
		stmt := &gorm.Statement{DB: m.DB}
		if err := stmt.Parse(oldName); err == nil {
			oldTable = stmt.Table
		} else {
			return err
		}
	}

	if v, ok := newName.(string); ok {
		newTable = v
	} else {
		stmt := &gorm.Statement{DB: m.DB}
		if err := stmt.Parse(newName); err == nil {
			newTable = stmt.Table
		} else {
			return err
		}
	}

	return m.DB.Exec("RENAME TABLE ? TO ?", clause.Table{Name: oldTable}, clause.Table{Name: newTable}).Error
}

func (m Migrator) DropColumn(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if field := stmt.Schema.LookUpField(name); field != nil {
			name = field.DBName
		}

		return m.DB.Exec(
			"ALTER TABLE ? DROP ?", clause.Table{Name: stmt.Table}, clause.Column{Name: name},
		).Error
	})
}

func (m Migrator) AlterColumn(value interface{}, field string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if field := stmt.Schema.LookUpField(field); field != nil {
			return m.DB.Exec(
				"ALTER TABLE ? MODIFY ? ?",
				clause.Table{Name: stmt.Table}, clause.Column{Name: field.DBName}, m.FullDataTypeOf(field),
			).Error
		}
		return fmt.Errorf("failed to look up field with name: %s", field)
	})
}

func (m Migrator) HasColumn(value interface{}, field string) bool {
	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		name := field
		if field := stmt.Schema.LookUpField(field); field != nil {
			name = field.DBName
		}

		return m.DB.Raw(
			"SELECT count(*) FROM USER_TAB_COLUMNS WHERE TABLE_NAME = UPPER(?) AND COLUMN_NAME = UPPER(?)",
			stmt.Table, name,
		).Row().Scan(&count)
	})

	return count > 0
}

func (m Migrator) CreateConstraint(value interface{}, name string) error {
	m.TryRemoveOnUpdate(value)
	return m.Migrator.CreateConstraint(value, name)
}

func (m Migrator) DropConstraint(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		for _, chk := range stmt.Schema.ParseCheckConstraints() {
			if chk.Name == name {
				return m.DB.Exec(
					"ALTER TABLE ? DROP CHECK ?",
					clause.Table{Name: stmt.Table}, clause.Column{Name: name},
				).Error
			}
		}

		return m.DB.Exec(
			"ALTER TABLE ? DROP CONSTRAINT ?",
			clause.Table{Name: stmt.Table}, clause.Column{Name: name},
		).Error
	})
}

func (m Migrator) HasConstraint(value interface{}, name string) bool {
	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		return m.DB.Raw(
			"SELECT count(*) FROM USER_CONSTRAINTS WHERE TABLE_NAME = UPPER(?) AND CONSTRAINT_NAME = UPPER(?)",
			stmt.Table, name,
		).Row().Scan(&count)
	})

	return count > 0
}

func (m Migrator) DropIndex(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if idx := stmt.Schema.LookIndex(name); idx != nil {
			name = idx.Name
		}

		return m.DB.Exec("DROP INDEX ?", clause.Column{Name: name}, clause.Table{Name: stmt.Table}).Error
	})
}

func (m Migrator) HasIndex(value interface{}, name string) bool {
	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if idx := stmt.Schema.LookIndex(name); idx != nil {
			name = idx.Name
		}

		return m.DB.Raw(
			"SELECT count(*) FROM USER_INDEXES WHERE TABLE_NAME = UPPER(?) AND INDEX_NAME = UPPER(?)",
			stmt.Table, name,
		).Row().Scan(&count)
	})

	return count > 0
}

// https://docs.oracle.com/database/121/SPATL/alter-index-rename.htm
func (m Migrator) RenameIndex(value interface{}, oldName, newName string) error {
	panic("TODO")
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		return m.DB.Exec(
			"ALTER INDEX ?.? RENAME TO ?", // wat
			clause.Table{Name: stmt.Table}, clause.Column{Name: oldName}, clause.Column{Name: newName},
		).Error
	})
}

func (m Migrator) TryRemoveOnUpdate(value interface{}) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		for _, rel := range stmt.Schema.Relationships.Relations {
			constraint := rel.ParseConstraint()
			if constraint != nil {
				rel.Field.TagSettings["CONSTRAINT"] = strings.ReplaceAll(rel.Field.TagSettings["CONSTRAINT"], fmt.Sprintf("ON UPDATE %s", constraint.OnUpdate), "")
			}
		}
		return nil
	})
}
