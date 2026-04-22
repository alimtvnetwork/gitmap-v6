// Package constants — constants_templates.go: keys for the templates feature
// (gitmap add ignore / attributes / lfs-install).
package constants

// Template kinds.
const (
	TemplateKindIgnore     = "ignore"
	TemplateKindAttributes = "attributes"
	TemplateKindLFS        = "lfs"
)

// User-templates directory name (joined under $HOME or %USERPROFILE%).
const (
	UserTemplatesDirName = ".gitmap"
	UserTemplatesSubdir  = "templates"
)

// Embedded asset paths (relative to gitmap/templates/embed.go).
const (
	EmbedAssetsRoot = "assets"
)

// File extensions used by templates.
const (
	TemplateExtIgnore     = ".gitignore"
	TemplateExtAttributes = ".gitattributes"
)

// Marker block tokens for managed regions inside .gitignore / .gitattributes.
const (
	MarkerIgnoreOpen     = "# >>> gitmap-ignore (do not edit between markers) >>>"
	MarkerIgnoreClose    = "# <<< gitmap-ignore <<<"
	MarkerAttributesOpen = "# >>> gitmap-attributes (do not edit between markers) >>>"
	MarkerAttributesClose = "# <<< gitmap-attributes <<<"
	MarkerUserEntries    = "# user entries"
)

// Template header field prefixes (used for audit-trail parsing).
const (
	TemplateHeaderSource  = "# source:"
	TemplateHeaderKind    = "# kind:"
	TemplateHeaderLang    = "# lang:"
	TemplateHeaderVersion = "# version:"
)

// Errors specific to the templates package.
const (
	ErrTemplateNotFound       = "template not found: kind=%s lang=%s"
	ErrTemplateMaterialize    = "failed to materialize templates to %s: %w"
	ErrTemplateUserDirResolve = "failed to resolve user templates dir: %w"
	ErrTemplateRead           = "failed to read template %s: %w"
)
