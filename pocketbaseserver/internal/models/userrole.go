package models

import (
	"github.com/pocketbase/pocketbase/core"
)

var _ core.RecordProxy = (*UserRole)(nil)

type UserRole struct {
	core.BaseRecordProxy
}

func (r *UserRole) User() string     { return r.GetString("user") }
func (r *UserRole) SetUser(v string) { r.Set("user", v) }
func (r *UserRole) Role() string     { return r.GetString("role") }
func (r *UserRole) SetRole(v string) { r.Set("role", v) }
