package oracle

import (
	"database/sql"

	"github.com/thoas/go-funk"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	gormSchema "gorm.io/gorm/schema"
)

func Create(db *gorm.DB) {
	stmt := db.Statement
	schema := stmt.Schema
	boundVars := make(map[string]int)

	if stmt == nil || schema == nil {
		return
	}

	if !stmt.Unscoped {
		for _, c := range schema.CreateClauses {
			stmt.AddClause(c)
		}
	}

	doExecution := func() {
		switch result, err := stmt.ConnPool.ExecContext(stmt.Context, stmt.SQL.String(), stmt.Vars...); err {
		case nil: // success
			db.RowsAffected, _ = result.RowsAffected()

			if len(schema.FieldsWithDefaultDBValue) > 0 {
				// bind returning back value to reflected value in the respective fields
				funk.ForEach(
					funk.Filter(schema.FieldsWithDefaultDBValue, func(field *gormSchema.Field) bool { return funk.Contains(boundVars, field.Name) }),
					func(field *gormSchema.Field) {
						field.Set(stmt.ReflectValue, stmt.Vars[boundVars[field.Name]].(sql.Out).Dest)
					},
				)
			}
		default: // failure
			db.AddError(err)
		}
	}

	if stmt.SQL.String() == "" {
		values := callbacks.ConvertToCreateValues(stmt)
		onConflict, hasConflict := stmt.Clauses["ON CONFLICT"].Expression.(clause.OnConflict)

		writeColumn := func() {
			for idx, column := range values.Columns {
				if idx > 0 {
					stmt.WriteByte(',')
				}
				stmt.WriteQuoted(column)
			}
		}
		genInsertIntoClause := func() {
			for _, value := range values.Values {
				WriteValues := func() {
					stmt.WriteString(" VALUES (")
					stmt.AddVar(stmt, value...)
					stmt.WriteByte(')')
				}
				stmt.WriteString(" INTO ")
				stmt.WriteString(stmt.Table)
				stmt.WriteByte('(')
				writeColumn()
				stmt.WriteByte(')')
				WriteValues()
			}
		}
		appendReturningIntoIfAppropriate := func() {
			doReturning := func() {
				for idx, field := range schema.FieldsWithDefaultDBValue {
					if idx > 0 {
						stmt.WriteByte(',')
					}
					stmt.WriteString(field.DBName)
				}
			}
			doInto := func() {
				for idx, field := range schema.FieldsWithDefaultDBValue {
					// fuck
					GenValueForReflection := func() interface{} {
						var value interface{}

						switch field.DataType {
						case gormSchema.Bool, gormSchema.Int, gormSchema.Uint:
							value = new(int)
						case gormSchema.Float:
							value = new(float64)
						case gormSchema.String:
							value = new(string)
						case gormSchema.Bytes:
							value = new([]byte)
						case gormSchema.Time:
							value = 0
						}
						return value
					}

					if idx > 0 {
						stmt.WriteByte(',')
					}

					boundVars[field.Name] = len(stmt.Vars)
					stmt.AddVar(stmt, sql.Out{Dest: GenValueForReflection()})
				}
			}
			if len(schema.FieldsWithDefaultDBValue) > 0 {
				stmt.WriteString(" RETURNING ")
				doReturning()
				stmt.WriteString(" INTO ")
				doInto()
			}
		}
		genInsertAllInto := func() {
			stmt.AddClauseIfNotExists(clause.Insert{Table: clause.Table{Name: stmt.Table}})
			stmt.WriteString("INSERT ALL")
			stmt.AddClause(values)
			if values, ok := stmt.Clauses["VALUES"].Expression.(clause.Values); ok {
				if len(values.Columns) > 0 {
					genInsertIntoClause()
					stmt.WriteString(" SELECT 1 FROM DUAL")
				}
				appendReturningIntoIfAppropriate()
			}
		}
		genMergeCreate := func() {
			createExcludedTableExpr := func() {
				var where clause.Where
				for _, field := range schema.PrimaryFields {
					where.Exprs = append(where.Exprs, clause.Eq{
						Column: clause.Column{Table: stmt.Table, Name: field.DBName},
						Value:  clause.Column{Table: "excluded", Name: field.DBName},
					})
				}
				where.Build(stmt)
			}
			doInsert := func() {
				stmt.WriteString(" INSERT (")
				InsertNames(values, schema, stmt, func(column clause.Column) interface{} {
					return column.Name
				})
				stmt.WriteByte(')')

				stmt.WriteString(" VALUES (")
				InsertNames(values, schema, stmt, func(column clause.Column) interface{} {
					return clause.Column{
						Table: "excluded",
						Name:  column.Name,
					}
				})
				stmt.WriteByte(')')
			}
			emulateMultiValuesRows := func() {
				for idx := 0; idx < len(values.Values); idx++ {
					if idx > 0 {
						stmt.WriteString(" UNION ALL")
					}
					stmt.WriteString(" SELECT ")
					for _, valCol := range funk.Zip(values.Values[idx], values.Columns) {
						stmt.AddVar(stmt, valCol.Element1)
						stmt.WriteString(" AS ")
						stmt.WriteQuoted(valCol.Element2)
					}
					stmt.WriteString(" FROM DUAL")
				}

			}

			stmt.WriteString("MERGE INTO ")
			stmt.WriteQuoted(stmt.Table)

			// Using
			stmt.WriteString(" USING (")
			emulateMultiValuesRows()
			stmt.WriteString(") excluded")

			// On condition
			stmt.WriteString(" ON (")
			createExcludedTableExpr()
			stmt.WriteByte(')')

			if len(onConflict.DoUpdates) > 0 {
				stmt.WriteString(" WHEN MATCHED THEN")
				stmt.WriteString(" UPDATE SET ")
				onConflict.DoUpdates.Build(stmt)
			}

			stmt.WriteString(" WHEN NOT MATCHED THEN")
			doInsert()

			appendReturningIntoIfAppropriate()
		}

		// are all columns in value the primary fields in schema only?
		if hasConflict && funk.Contains(
			funk.Map(values.Columns, func(c clause.Column) string { return c.Name }),
			funk.Map(schema.PrimaryFields, func(field *gormSchema.Field) string { return field.DBName }),
		) {
			genMergeCreate()
		} else {
			genInsertAllInto()
		}
	}

	if !db.DryRun {
		doExecution()
	}
}

func InsertNames(values clause.Values, schema *gormSchema.Schema, stmt *gorm.Statement, valueFn func(column clause.Column) interface{}) {
	for idx, column := range values.Columns {
		if !(schema.PrioritizedPrimaryField != nil && schema.PrioritizedPrimaryField.AutoIncrement && schema.PrioritizedPrimaryField.DBName == column.Name) {
			if idx > 0 {
				stmt.WriteByte(',')
			}
			stmt.WriteQuoted(valueFn(column))
		}
	}
}
