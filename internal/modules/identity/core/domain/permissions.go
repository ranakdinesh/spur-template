package domain

import (
	"github.com/spurbase/spur/internal/platform/appmodule"
)

// Permissions Registry
// Usage: identity.Permissions.Tenants.Create
var Permissions = struct {
	Tenants appmodule.ResourcePerms
	Users   appmodule.ResourcePerms
	Roles   appmodule.ResourcePerms
	APIKeys appmodule.ResourcePerms
	Modules appmodule.ResourcePerms
}{
	Tenants: appmodule.NewResourcePerms(ModuleCode, "tenants"),
	Users:   appmodule.NewResourcePerms(ModuleCode, "users"),
	Roles:   appmodule.NewResourcePerms(ModuleCode, "roles"),
	APIKeys: appmodule.NewResourcePerms(ModuleCode, "apikeys"),
	Modules: appmodule.NewResourcePerms(ModuleCode, "modules"),
}
