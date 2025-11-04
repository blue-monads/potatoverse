package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")

}

type DocSpec struct {
	Title       string      `json:"title" toml:"title"`
	Description string      `json:"description" toml:"description"`
	Methods     []DocMethod `json:"methods" toml:"methods"`
}

type DocMethod struct {
	Method      string         `json:"method" toml:"method"`
	Path        string         `json:"path" toml:"path"`
	Description string         `json:"description" toml:"description"`
	Parameters  []DocParameter `json:"parameters" toml:"parameters"`
	Responses   []DocResponse  `json:"responses" toml:"responses"`
}

type DocParameter struct {
	Name        string `json:"name" toml:"name"`
	Description string `json:"description" toml:"description"`
	Type        string `json:"type" toml:"type"`
	Required    bool   `json:"required" toml:"required"`
}

type DocResponse struct {
	Code        int    `json:"code" toml:"code"`
	Description string `json:"description" toml:"description"`
	Schema      any    `json:"schema" toml:"schema"`
}
