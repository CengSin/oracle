package clauses

import (
	"gorm.io/gorm/clause"
)

type WhenMathced struct {
	clause.Set
	Where, Delete clause.Where
}

type WhenNotMatched struct {
	clause.Values
	Where clause.Where
}

type Merge struct {
	Table          clause.Table
	Using          []clause.Interface
	On             []clause.Expression
	WhenMatched    WhenMathced
	WhenNotMatched WhenNotMatched
}

func (merge Merge) Name() string {
	return "MERGE"
}

func MergeDefaultExcludeName() string {
	return "exclude"
}

// Build build from clause
func (merge Merge) Build(builder clause.Builder) {
	builder.WriteString("INTO ")
	if merge.Table.Name == "" {
		builder.WriteQuoted(clause.Table{Name: clause.CurrentTable})
	} else {
		builder.WriteQuoted(merge.Table)
	}
	builder.WriteString(" USING (")
	for idx, iface := range merge.Using {
		if idx > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(iface.Name())
		builder.WriteByte(' ')
		iface.Build(builder)
	}
	builder.WriteString(") ")
	builder.WriteString(MergeDefaultExcludeName())
	builder.WriteString(" ON (")
	for idx, on := range merge.On {
		if idx > 0 {
			builder.WriteString(", ")
		}
		on.Build(builder)
	}
	builder.WriteString(")")

	if len(merge.WhenMatched.Set) > 0 {
		pred := merge.WhenMatched
		builder.WriteString(" WHEN MATCHED THEN")
		builder.WriteString(" UPDATE ")
		builder.WriteString(pred.Name())
		builder.WriteByte(' ')
		pred.Build(builder)

		buildWhere := func(where clause.Where) {
			builder.WriteString(where.Name())
			builder.WriteByte(' ')
			where.Build(builder)
		}

		if len(pred.Where.Exprs) > 0 {
			buildWhere(pred.Where)
		}

		if len(pred.Delete.Exprs) > 0 {
			builder.WriteString(" DELETE ")
			buildWhere(pred.Delete)
		}
	}

	if len(merge.WhenNotMatched.Columns) > 0 {
		pred := merge.WhenNotMatched
		if len(pred.Values.Values) != 1 {
			panic("cannot insert more than one rows due to Oracle SQL language restriction")
		}

		builder.WriteString(" WHEN NOT MATCHED THEN")
		builder.WriteString(" INSERT ")
		pred.Build(builder)

		if len(pred.Where.Exprs) > 0 {
			builder.WriteString(pred.Where.Name())
			builder.WriteByte(' ')
			pred.Where.Build(builder)
		}
	}
}

// MergeClause merge values clauses
func (merge Merge) MergeClause(clause *clause.Clause) {
	clause.Name = merge.Name()
	clause.Expression = merge
}
