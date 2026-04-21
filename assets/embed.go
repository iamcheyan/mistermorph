package assets

import "embed"

// SkillsFS contains built-in skills shipped with mistermorph (under assets/skills).
//
//go:embed skills/**
var SkillsFS embed.FS

// ConfigFS contains embedded config templates shipped with mistermorph (under assets/config).
//
//go:embed config/**
var ConfigFS embed.FS

// CreditsFS contains embedded credits data shipped with mistermorph (under assets/credits).
//
//go:embed credits/**
var CreditsFS embed.FS
