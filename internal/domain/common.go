package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

type ID string

func (id ID) Valid() bool {
	v := strings.TrimSpace(string(id))
	return len(v) >= 3 && len(v) <= 80 && !strings.ContainsAny(v, " /\\")
}

type Actor struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
	Role Role   `json:"role"`
}

type Role string

const (
	RoleAdmin      Role = "admin"
	RolePlanner    Role = "planner"
	RoleMaintainer Role = "maintainer"
	RoleReviewer   Role = "reviewer"
	RoleViewer     Role = "viewer"
)

func (a Actor) CanPlan() bool { return a.Role == RoleAdmin || a.Role == RolePlanner }
func (a Actor) CanExecute() bool {
	return a.Role == RoleAdmin || a.Role == RoleMaintainer
}
func (a Actor) CanReview() bool { return a.Role == RoleAdmin || a.Role == RoleReviewer }

type Page struct {
	Offset int    `form:"offset" json:"offset"`
	Limit  int    `form:"limit" json:"limit"`
	Sort   string `form:"sort" json:"sort"`
}

func (p Page) Normalize(allowed map[string]bool) Page {
	if p.Offset < 0 {
		p.Offset = 0
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	if !allowed[p.Sort] {
		p.Sort = "created_at"
	}
	return p
}

func StableKey(parts ...string) string {
	h := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(h[:16])
}

func UTCDate(t time.Time) string { return t.UTC().Format("2006-01-02") }

type Versioned struct {
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
