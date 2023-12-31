package hints

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Hints struct {
	HintMsg string
	Prefix  string
	Suffix  string
	Content string

	clause string
	before bool
	after  bool
}

func (hints Hints) ModifyStatement(stmt *gorm.Statement) {
	name := strings.ToUpper(hints.clause)
	if name == "" {
		name = "SELECT"
	}

	clause := stmt.Clauses[name]
	switch {
	case hints.before:
		if clause.BeforeExpression == nil {
			clause.BeforeExpression = hints
		} else if old, ok := clause.BeforeExpression.(Hints); ok {
			old.Merge(hints)
			clause.BeforeExpression = old
		} else {
			clause.BeforeExpression = Exprs{clause.BeforeExpression, hints}
		}
	case hints.after:
		if clause.AfterExpression == nil {
			clause.AfterExpression = hints
		} else if old, ok := clause.AfterExpression.(Hints); ok {
			old.Merge(hints)
			clause.AfterExpression = old
		} else {
			clause.AfterExpression = Exprs{clause.AfterExpression, hints}
		}
	default:
		if clause.AfterNameExpression == nil {
			clause.AfterNameExpression = hints
		} else if old, ok := clause.AfterNameExpression.(Hints); ok {
			old.Merge(hints)
			clause.AfterNameExpression = old
		} else {
			clause.AfterNameExpression = Exprs{clause.AfterNameExpression, hints}
		}
	}

	stmt.Clauses[name] = clause
}

func (hints Hints) Build(builder clause.Builder) {
	if hints.HintMsg != "" {
		builder.WriteString(hints.HintMsg)
	} else {
		builder.WriteString(hints.Prefix)
		builder.WriteString(hints.Content)
		builder.WriteString(hints.Suffix)
	}
}

func (hints Hints) Merge(h Hints) {
	hints.Content += " " + h.Content
}

func New(content string) Hints {
	return Hints{Prefix: "/*+ ", Content: content, Suffix: " */"}
}

func NewHint(msg string) Hints {
	return Hints{HintMsg: msg}
}

func Comment(clause string, comment string) Hints {
	return Hints{clause: clause, Prefix: "/* ", Content: comment, Suffix: " */"}
}

func CommentBefore(clause string, comment string) Hints {
	return Hints{clause: clause, before: true, Prefix: "/* ", Content: comment, Suffix: " */"}
}

func CommentAfter(clause string, comment string) Hints {
	return Hints{clause: clause, after: true, Prefix: "/* ", Content: comment, Suffix: " */"}
}
