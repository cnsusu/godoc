<p align="right">
   <a href="./README.md">中文</a> | <strong>English</strong>
</p>

# Swagger to Markdown Generator (Go)
[![Go Report Card](https://goreportcard.com/badge/github.com/cnsusu/swaggermarkdown)](https://goreportcard.com/report/github.com/cnsusu/swaggermarkdown)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

> A lightweight, dependency-free Go library for converting Swagger/OpenAPI JSON specifications into well-structured Markdown documentation with customizable formatting options. Perfect for static site deployment and offline usage.

## ✨ Features
- **Dynamic Order Control** - Customize endpoint display order via priority mapping
- **Field Filtering** - Exclude sensitive or unnecessary fields (e.g. `_app_id`)
- **Single-File Output** - Generate consolidated Markdown files
- **Zero Dependencies** - Pure Go implementation without external dependencies
- **Swagger 2.0 & OpenAPI 3.0** - Full Markdown formatting support including tables, code blocks, and links :cite[5]

## 📥 Installation
```bash
go get github.com/cnsusu/swaggermarkdow
```

## 🚀 Basic Usage
```bash
package main

import (
	"log"
	swaggermarkdown "github.com/cnsusu/swaggermarkdown"
)

func main() {
	// Define fields to ignore in the output
	ignoredFields := map[string]bool{
		"_app_id":    true,  // Sensitive app ID
		"internal_id": true, // Add other fields here
	}

	// Configure endpoint display order (lower = earlier)
	customOrder := map[string]int{
		"/api/user/register": 1, // Highest priority
		"/api/user/login":    2,
	}

	// Initialize generator
	swaggerMarkdown := swaggermarkdown.NewSwaggerMarkdown()
	swaggerMarkdown.SetOrder(customOrder)
	swaggerMarkdown.SetIgnored(ignoredFields)

	// Generate documentation
	err := swaggerMarkdown.GenerateFile("swagger.json", "API_Documentation.md")
	if err != nil {
		log.Fatal("Generation failed: ", err)
	}

}
```

