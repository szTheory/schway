package core

import "github.com/codename-lang/lang/internal/compiler/diagnostic"

const Schema = "lang.core/0"

type Program struct {
	Schema    string     `json:"schema"`
	Module    string     `json:"module"`
	ModuleID  string     `json:"module_id"`
	DataTypes []DataType `json:"data_types"`
	Functions []Function `json:"functions"`
}

type DataType struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Alternatives []string        `json:"alternatives"`
	Span         diagnostic.Span `json:"span"`
}

type Function struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Parameter  Parameter       `json:"parameter"`
	ReturnType string          `json:"return_type"`
	Match      Match           `json:"match"`
	Span       diagnostic.Span `json:"span"`
}

type Parameter struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Match struct {
	ID        string     `json:"id"`
	Scrutinee string     `json:"scrutinee"`
	Arms      []MatchArm `json:"arms"`
}

type MatchArm struct {
	ID      string `json:"id"`
	Pattern string `json:"pattern"`
	Value   string `json:"value"`
}
