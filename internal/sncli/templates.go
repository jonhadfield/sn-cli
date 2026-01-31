package sncli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Template represents a note template
type Template struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Title       string            `yaml:"title"`
	Content     string            `yaml:"content"`
	Tags        []string          `yaml:"tags"`
	Variables   map[string]string `yaml:"variables"`
}

// TemplateConfig holds template processing configuration
type TemplateConfig struct {
	TemplateDir string
	Variables   map[string]string
}

// BuiltInTemplates returns a map of built-in templates
func BuiltInTemplates() map[string]Template {
	return map[string]Template{
		"meeting": {
			Name:        "meeting",
			Description: "Meeting notes template",
			Title:       "Meeting: {{title}} - {{date}}",
			Content: `# Meeting Notes
**Date:** {{date}}
**Time:** {{time}}
**Attendees:** {{attendees}}

## Agenda
-

## Discussion Points
-

## Action Items
- [ ]

## Next Steps
-

## Notes
`,
			Tags: []string{"meetings"},
		},
		"daily": {
			Name:        "daily",
			Description: "Daily log template",
			Title:       "Daily Log - {{date}}",
			Content: `# Daily Log - {{date}}

## 🎯 Today's Goals
-

## ✅ Completed
-

## 📝 Notes
-

## 💡 Ideas
-

## 🔄 Tomorrow
-
`,
			Tags: []string{"daily", "journal"},
		},
		"todo": {
			Name:        "todo",
			Description: "Todo list template",
			Title:       "Todo: {{title}}",
			Content: `# {{title}}

## High Priority
- [ ]

## Medium Priority
- [ ]

## Low Priority
- [ ]

## Completed
- [x]
`,
			Tags: []string{"todo", "tasks"},
		},
		"project": {
			Name:        "project",
			Description: "Project planning template",
			Title:       "Project: {{title}}",
			Content: `# Project: {{title}}

**Created:** {{date}}
**Status:** Planning

## Overview
- **Goal:**
- **Timeline:**
- **Owner:** {{user}}

## Objectives
1.

## Milestones
- [ ] Phase 1:
- [ ] Phase 2:
- [ ] Phase 3:

## Resources
-

## Risks
-

## Notes
`,
			Tags: []string{"projects"},
		},
		"research": {
			Name:        "research",
			Description: "Research notes template",
			Title:       "Research: {{title}}",
			Content: `# Research: {{title}}

**Date:** {{date}}
**Source:**

## Summary
-

## Key Points
1.

## Questions
-

## References
-

## Follow-up
-
`,
			Tags: []string{"research"},
		},
		"idea": {
			Name:        "idea",
			Description: "Idea capture template",
			Title:       "💡 {{title}}",
			Content: `# 💡 {{title}}

**Date:** {{datetime}}

## The Idea
-

## Why It Matters
-

## Potential Applications
-

## Next Steps
- [ ]

## Related Ideas
-
`,
			Tags: []string{"ideas"},
		},
		"book": {
			Name:        "book",
			Description: "Book notes template",
			Title:       "📚 {{title}}",
			Content: `# 📚 {{title}}

**Author:**
**Date Read:** {{date}}
**Rating:** ⭐⭐⭐⭐⭐

## Summary
-

## Key Takeaways
1.

## Favorite Quotes
>

## My Thoughts
-

## Action Items
- [ ]
`,
			Tags: []string{"books", "reading"},
		},
		"retrospective": {
			Name:        "retrospective",
			Description: "Sprint/project retrospective template",
			Title:       "Retrospective: {{title}} - {{date}}",
			Content: `# Retrospective: {{title}}

**Date:** {{date}}
**Period:**

## 🟢 What Went Well
-

## 🔴 What Could Be Improved
-

## 💡 Ideas & Insights
-

## 🎯 Action Items
- [ ]

## 📊 Metrics
-
`,
			Tags: []string{"retrospective", "review"},
		},
	}
}

// ProcessTemplate processes a template with variable substitution
func ProcessTemplate(template Template, vars map[string]string) (title, content string, tags []string) {
	// Merge built-in variables with custom ones
	allVars := GetDefaultVariables()
	for k, v := range vars {
		allVars[k] = v
	}

	// Process title
	title = template.Title
	for key, val := range allVars {
		title = strings.ReplaceAll(title, "{{"+key+"}}", val)
	}

	// Process content
	content = template.Content
	for key, val := range allVars {
		content = strings.ReplaceAll(content, "{{"+key+"}}", val)
	}

	// Return tags
	tags = template.Tags

	return title, content, tags
}

// GetDefaultVariables returns default template variables
func GetDefaultVariables() map[string]string {
	now := time.Now()

	return map[string]string{
		"date":     now.Format("2006-01-02"),
		"time":     now.Format("15:04"),
		"datetime": now.Format("2006-01-02 15:04:05"),
		"year":     now.Format("2006"),
		"month":    now.Format("01"),
		"day":      now.Format("02"),
		"weekday":  now.Format("Monday"),
		"user":     os.Getenv("USER"),
	}
}

// GetTemplate retrieves a template by name (built-in or custom)
func GetTemplate(name string, templateDir string) (Template, error) {
	// Check built-in templates first
	if tpl, ok := BuiltInTemplates()[name]; ok {
		return tpl, nil
	}

	// Check custom templates
	if templateDir == "" {
		return Template{}, fmt.Errorf("template '%s' not found", name)
	}

	templatePath := filepath.Join(templateDir, name+".yaml")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		templatePath = filepath.Join(templateDir, name+".yml")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			return Template{}, fmt.Errorf("template '%s' not found", name)
		}
	}

	// Load custom template
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return Template{}, fmt.Errorf("failed to read template: %w", err)
	}

	var tpl Template
	if err := yaml.Unmarshal(data, &tpl); err != nil {
		return Template{}, fmt.Errorf("failed to parse template: %w", err)
	}

	return tpl, nil
}

// ListTemplates returns all available templates
func ListTemplates(templateDir string) []Template {
	var templates []Template

	// Add built-in templates
	for _, tpl := range BuiltInTemplates() {
		templates = append(templates, tpl)
	}

	// Add custom templates if directory exists
	if templateDir != "" {
		if entries, err := os.ReadDir(templateDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				name := entry.Name()
				if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
					templateName := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")

					// Skip if already in built-ins
					if _, ok := BuiltInTemplates()[templateName]; ok {
						continue
					}

					// Load custom template
					if tpl, err := GetTemplate(templateName, templateDir); err == nil {
						templates = append(templates, tpl)
					}
				}
			}
		}
	}

	return templates
}

// SaveTemplate saves a template to the template directory
func SaveTemplate(template Template, templateDir string) error {
	if templateDir == "" {
		return fmt.Errorf("template directory not specified")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	// Save template
	templatePath := filepath.Join(templateDir, template.Name+".yaml")
	data, err := yaml.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	if err := os.WriteFile(templatePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	return nil
}

// GetTemplateDir returns the default template directory
func GetTemplateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "sn-cli", "templates"), nil
}
