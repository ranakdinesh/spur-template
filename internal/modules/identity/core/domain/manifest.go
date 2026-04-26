package domain

// Manifest declares the permissions a module provides.
// Every module constructs one in New() and passes it to
// identityModule.Services.ModuleService.RegisterManifest().
type Manifest struct {
	Name        string
	Code        string // unique slug, e.g. "leadcrm"
	Description string
	Permissions []ManifestPermission
}

// ManifestPermission is a permission declaration inside a Manifest.
// Slug is the full dot-separated permission string: "leadcrm.leads.view"
type ManifestPermission struct {
	Slug        string // e.g. "leadcrm.leads.view"
	Description string
}
