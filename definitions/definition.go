package definitions

import "github.com/jacksonzamorano/strata/component"

var Manifest = component.ComponentManifest{
	Name:    "strata-search",
	Version: "1.0.0",
}

type SearchInput struct {
	Path    string
	Pattern string
	Context int
}
type SearchOutput struct {
	Matches     []SearchMatch
	TimeSeconds float64
}
type SearchMatch struct {
	File         string
	AbsolutePath string

	Before string
	Match  string
	After  string

	LineNumber int
}

var Search = component.Define[SearchInput, SearchOutput](Manifest, "search")
