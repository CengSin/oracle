package clauses

import (
	"gorm.io/gorm/clause"
)

type Merge struct {
	Table clause.Table
	Using []clause.Interface
	On    []clause.Expression
}

func (merge Merge) Name() string {
	return "MERGE"
}

func MergeDefaultExcludeName() string {
	return "exclude"
}

// Build build from clause
func (merge Merge) Build(builder clause.Builder) {
	clause.Insert{}.Build(builder)
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
}

// MergeClause merge values clauses
func (merge Merge) MergeClause(clause *clause.Clause) {
	clause.Name = merge.Name()
	clause.Expression = merge
}
