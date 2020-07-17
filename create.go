package oracle

import (
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
)

func Create(db *gorm.DB) {
	stmt := db.Statement
	schema := stmt.Schema

	if schema == nil {
		return
	}

	if !stmt.Unscoped {
		for _, c := range schema.CreateClauses {
			stmt.AddClause(c)
		}
	}

	var (
		boundVars map[string]int
	)

	if stmt.SQL.String() == "" {
		var (
			values                  = callbacks.ConvertToCreateValues(stmt)
			c                       = stmt.Clauses["ON CONFLICT"]
			onConflict, hasConflict = c.Expression.(clause.OnConflict)
		)

		if hasConflict {
			if len(schema.PrimaryFields) > 0 {
				columnsMap := map[string]bool{}
				for _, column := range values.Columns {
					columnsMap[column.Name] = true
				}

				for _, field := range schema.PrimaryFields {
					if _, ok := columnsMap[field.DBName]; !ok {
						hasConflict = false
					}
				}
			} else {
				hasConflict = false
			}
		}

		if hasConflict {
			boundVars = MergeCreate(db, onConflict, values)
		} else {
			stmt.AddClauseIfNotExists(clause.Insert{Table: clause.Table{Name: stmt.Table}})
			stmt.Build("INSERT")
			stmt.WriteByte(' ')

			stmt.AddClause(values)
			if values, ok := stmt.Clauses["VALUES"].Expression.(clause.Values); ok {
				if len(values.Columns) > 0 {
					stmt.WriteByte('(')
					for idx, column := range values.Columns {
						if idx > 0 {
							stmt.WriteByte(',')
						}
						stmt.WriteQuoted(column)
					}
					stmt.WriteByte(')')

					stmt.WriteString(" VALUES ")

					for idx, value := range values.Values {
						if idx > 0 {
							stmt.WriteByte(',')
						}

						stmt.WriteByte('(')
						stmt.AddVar(stmt, value...)
						stmt.WriteByte(')')
					}
				}
				boundVars = outputInserted(db)
			}
		}
	}

	if !db.DryRun {
		if result, err := stmt.ConnPool.ExecContext(stmt.Context, stmt.SQL.String(), stmt.Vars...); err == nil {
			db.RowsAffected, _ = result.RowsAffected()

			if len(schema.FieldsWithDefaultDBValue) > 0 {
				for _, field := range schema.FieldsWithDefaultDBValue {
					field.Set(stmt.ReflectValue, stmt.Vars[boundVars[field.Name]].(sql.Out).Dest)
				}
			}
		} else {
			db.AddError(err)
		}
	}
}

func MergeCreate(db *gorm.DB, onConflict clause.OnConflict, values clause.Values) (boundVars map[string]int) {
	stmt := db.Statement
	stmt.WriteString("MERGE INTO ")
	stmt.WriteQuoted(stmt.Table)
	stmt.WriteString(" USING (")

	for idx, column := range values.Columns {
		stmt.WriteString(" SELECT")
		stmt.AddVar(stmt, values.Values[idx]...)
		stmt.WriteString(" AS ")
		stmt.WriteQuoted(column.Name)
		stmt.WriteString(" FROM DUAL")
		stmt.WriteString(" UNION ALL")
	}

	stmt.WriteString(") excluded ON ")

	var where clause.Where
	schema := stmt.Schema
	for _, field := range schema.PrimaryFields {
		where.Exprs = append(where.Exprs, clause.Eq{
			Column: clause.Column{Table: stmt.Table, Name: field.DBName},
			Value:  clause.Column{Table: "excluded", Name: field.DBName},
		})
	}
	where.Build(stmt)

	if len(onConflict.DoUpdates) > 0 {
		stmt.WriteString(" WHEN MATCHED THEN UPDATE SET ")
		onConflict.DoUpdates.Build(stmt)
	}

	stmt.WriteString(" WHEN NOT MATCHED THEN INSERT (")

	written := false
	for _, column := range values.Columns {
		if schema.PrioritizedPrimaryField == nil || !schema.PrioritizedPrimaryField.AutoIncrement || schema.PrioritizedPrimaryField.DBName != column.Name {
			if written {
				stmt.WriteByte(',')
			}
			written = true
			stmt.WriteQuoted(column.Name)
		}
	}

	stmt.WriteString(") VALUES (")

	written = false
	for _, column := range values.Columns {
		if !(schema.PrioritizedPrimaryField != nil && schema.PrioritizedPrimaryField.AutoIncrement && schema.PrioritizedPrimaryField.DBName == column.Name) {
			if written {
				stmt.WriteByte(',')
			}
			written = true
			stmt.WriteQuoted(clause.Column{
				Table: "excluded",
				Name:  column.Name,
			})
		}
	}

	stmt.WriteString(")")
	return outputInserted(db)
}

func outputInserted(db *gorm.DB) (boundVars map[string]int) {
	stmt := db.Statement
	schema := stmt.Schema
	if len(schema.FieldsWithDefaultDBValue) > 0 {
		stmt.WriteString(" RETURNING ")

		boundVars = make(map[string]int)
		for idx, field := range schema.FieldsWithDefaultDBValue {
			if idx > 0 {
				stmt.WriteString(",")
			}
			stmt.WriteString(field.DBName)
		}

		stmt.WriteString(" INTO ")
		for idx, field := range schema.FieldsWithDefaultDBValue {
			if idx > 0 {
				stmt.WriteString(",")
			}

			boundVars[field.Name] = len(stmt.Vars)
			stmt.AddVar(stmt, sql.Out{Dest: field.ReflectValueOf(stmt.ReflectValue).Addr().Interface()})
		}
	}
	return
}
