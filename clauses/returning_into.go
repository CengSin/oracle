package clauses

import (
	"gorm.io/gorm/clause"
)

type ReturningInto struct {
	Variables []clause.Column
	Into      []*clause.Values
}
