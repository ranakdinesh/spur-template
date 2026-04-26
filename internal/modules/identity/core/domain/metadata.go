package domain

import (
	"github.com/spurbase/spur/internal/platform/appmodule"
)

const (
	ModuleCode = "identity"
	ModuleName = "Identity & Access Management"
	ModuleDesc = "Handles authentication, authorization, tenants, and users."
)

// Permission Usage Helpers
// Instead of 100 constants, we can define them structurally if needed, 
// or keep key constants for specific usage.
// For now, I'll keep the custom ones and let the resource ones be generated dynamically in DB.
// To use them in code: `identity.Perm("users", "list")` or define specific needed constants.

const (
	PermAdminAccess = "identity.admin.access"
)

func GetModuleDefinition() appmodule.Definition {
	return appmodule.NewBuilder(ModuleCode, ModuleName, ModuleDesc).
		// 1. Custom Single Permissions
		AddPermission("admin.access", "Access to Identity Admin dashboard").
		
		// 2. Resource Groups (Auto-generates CRUD permissions)
		// Generates: tenants.list, tenants.create, tenants.view, tenants.update, tenants.delete
		AddResource("tenants", "Tenant", appmodule.CRUD...).
		
		AddResource("users", "User", appmodule.CRUD...).
		
		AddResource("roles", "Role", appmodule.CRUD...).
		
		AddResource("apikeys", "API Key", appmodule.ActionList, appmodule.ActionCreate, appmodule.ActionDelete).
		
		AddResource("modules", "System Module", appmodule.CRUD...).
		
		Build()
}