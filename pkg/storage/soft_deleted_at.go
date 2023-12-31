package storage

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type DeletedAt sql.NullTime

// Scan implements the Scanner interface.
func (n *DeletedAt) Scan(value interface{}) error {
	return (*sql.NullTime)(n).Scan(value)
}

// Value implements the driver Valuer interface.
func (n DeletedAt) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time, nil
}

func (n DeletedAt) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Time)
	}
	return json.Marshal(nil)
}

func (n *DeletedAt) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		n.Valid = false
		return nil
	}
	err := json.Unmarshal(b, &n.Time)
	if err == nil {
		n.Valid = true
	}
	return err
}

func (d DeletedAt) UpdateClauses(field *schema.Field) []clause.Interface {
	return []clause.Interface{SoftDeletedAtUpdateClause{Field: field}}
}

type SoftDeletedAtUpdateClause struct {
	Field *schema.Field
}

func (s SoftDeletedAtUpdateClause) Name() string {
	return ""
}

func (s SoftDeletedAtUpdateClause) Build(clause.Builder) {
}

func (s SoftDeletedAtUpdateClause) MergeClause(*clause.Clause) {
}

func (s SoftDeletedAtUpdateClause) ModifyStatement(stmt *gorm.Statement) {
	if _, ok := stmt.Clauses["soft_delete_enabled"]; ok || stmt.Statement.Unscoped {
		return
	}
	if c, ok := stmt.Clauses["WHERE"]; ok {
		if where, ok := c.Expression.(clause.Where); ok && len(where.Exprs) > 1 {
			for _, expr := range where.Exprs {
				if orCond, ok := expr.(clause.OrConditions); ok && len(orCond.Exprs) == 1 {
					where.Exprs = []clause.Expression{clause.And(where.Exprs...)}
					c.Expression = where
					stmt.Clauses["WHERE"] = c
					break
				}
			}
		}
	}
	stmt.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Eq{Column: clause.Column{Table: clause.CurrentTable, Name: s.Field.DBName}, Value: nil},
	}})
	stmt.Clauses["soft_delete_enabled"] = clause.Clause{}
}

func (DeletedAt) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{SoftDeleteQueryClause{Field: f}}
}

type SoftDeleteQueryClause struct {
	Field *schema.Field
}

func (sd SoftDeleteQueryClause) Name() string {
	return ""
}

func (sd SoftDeleteQueryClause) Build(clause.Builder) {
}

func (sd SoftDeleteQueryClause) MergeClause(*clause.Clause) {
}

func (sd SoftDeleteQueryClause) ModifyStatement(stmt *gorm.Statement) {
	if _, ok := stmt.Clauses["soft_delete_enabled"]; ok || stmt.Statement.Unscoped {
		return
	}
	if c, ok := stmt.Clauses["WHERE"]; ok {
		if where, ok := c.Expression.(clause.Where); ok && len(where.Exprs) > 1 {
			for _, expr := range where.Exprs {
				if orCond, ok := expr.(clause.OrConditions); ok && len(orCond.Exprs) == 1 {
					where.Exprs = []clause.Expression{clause.And(where.Exprs...)}
					c.Expression = where
					stmt.Clauses["WHERE"] = c
					break
				}
			}
		}
	}

	stmt.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Eq{Column: clause.Column{Table: clause.CurrentTable, Name: sd.Field.DBName}, Value: nil},
	}})
	stmt.Clauses["soft_delete_enabled"] = clause.Clause{}
}

func (DeletedAt) DeleteClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{SoftDeleteDeleteClause{Field: f}}
}

type SoftDeleteDeleteClause struct {
	Field *schema.Field
}

func (sd SoftDeleteDeleteClause) Name() string {
	return ""
}

func (sd SoftDeleteDeleteClause) Build(clause.Builder) {
}

func (sd SoftDeleteDeleteClause) MergeClause(*clause.Clause) {
}

func (sd SoftDeleteDeleteClause) ModifyStatement(stmt *gorm.Statement) {
	if stmt.SQL.Len() != 0 || stmt.Statement.Unscoped {
		return
	}
	var clauseSet clause.Set
	curTime := stmt.NowFunc()
	clauseSet = append(clauseSet, clause.Assignment{Column: clause.Column{Name: sd.Field.DBName}, Value: curTime})

	if _, ok := stmt.Schema.FieldsByName["Deleted"]; ok {
		assignment := clause.Assignment{
			Column: clause.Column{Name: "deleted"},
			Value:  gorm.Expr("`id`"),
		}
		_, idOk := stmt.Schema.FieldsByDBName["id"]
		if !idOk {
			assignment.Value = uint64(curTime.UnixNano())
			return
		}
		clauseSet = append(clauseSet, assignment)
	}
	if filed, ok := stmt.Schema.FieldsByName["Status"]; ok {
		if !strings.Contains(filed.Comment, "skip_delete") {
			clauseSet = append(clauseSet, clause.Assignment{Column: clause.Column{Name: "status"}, Value: "deleted"})
		}
	}
	stmt.AddClause(clauseSet)
	if stmt.Schema != nil {
		_, queryValues := schema.GetIdentityFieldValuesMap(stmt.Context, stmt.ReflectValue, stmt.Schema.PrimaryFields)
		column, values := schema.ToQueryValues(stmt.Table, stmt.Schema.PrimaryFieldDBNames, queryValues)

		if len(values) > 0 {
			stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
		}

		if stmt.ReflectValue.CanAddr() && stmt.Dest != stmt.Model && stmt.Model != nil {
			_, queryValues = schema.GetIdentityFieldValuesMap(stmt.Context, reflect.ValueOf(stmt.Model), stmt.Schema.PrimaryFields)
			column, values = schema.ToQueryValues(stmt.Table, stmt.Schema.PrimaryFieldDBNames, queryValues)

			if len(values) > 0 {
				stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.IN{Column: column, Values: values}}})
			}
		}
	}

	if _, ok := stmt.Clauses["WHERE"]; !stmt.AllowGlobalUpdate && !ok {
		_ = stmt.AddError(gorm.ErrMissingWhereClause)
	} else {
		SoftDeleteQueryClause(sd).ModifyStatement(stmt)
	}

	stmt.AddClauseIfNotExists(clause.Update{})
	stmt.Build("UPDATE", "SET", "WHERE")
}
