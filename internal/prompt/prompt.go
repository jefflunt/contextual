package prompt

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/jluntpcty/contextual/internal/types"
)

//go:embed writing-plan-files.md
var planFileSpec string

// BuildPlanPrompt assembles the full prompt to pass to copilot.
// contextualFile is the path to the contextual output file.
// primaryItem is the first item fetched (the main subject).
func BuildPlanPrompt(contextualFile string, primaryItem types.Item) string {
	var sb strings.Builder

	sb.WriteString("You are an expert software engineering assistant. ")
	sb.WriteString("Your task is to read a context file and produce a complete plan file ")
	sb.WriteString("for the primary work item it describes.\n\n")

	sb.WriteString("---\n\n")
	sb.WriteString("## Plan File Specification\n\n")
	sb.WriteString("The following is the full specification for how to write a plan file. ")
	sb.WriteString("Follow it exactly.\n\n")
	sb.WriteString(planFileSpec)
	sb.WriteString("\n\n---\n\n")

	sb.WriteString("## Context File\n\n")
	sb.WriteString(fmt.Sprintf(
		"The context file is located at: `%s`\n\n", contextualFile,
	))
	sb.WriteString("Read this file in full before writing the plan. ")
	sb.WriteString("It contains one or more fetched items (Jira issues, Confluence pages, web pages). ")

	switch primaryItem.Type {
	case types.ItemTypeJira:
		sb.WriteString(fmt.Sprintf(
			"The **primary subject** is the Jira issue **%s** (\"%s\"). ",
			primaryItem.ID, primaryItem.Title,
		))
	case types.ItemTypeConfluence:
		sb.WriteString(fmt.Sprintf(
			"The **primary subject** is the Confluence page **%s** (\"%s\"). ",
			primaryItem.ID, primaryItem.Title,
		))
	case types.ItemTypeWeb:
		sb.WriteString(fmt.Sprintf(
			"The **primary subject** is the web page at **%s** (\"%s\"). ",
			primaryItem.URL, primaryItem.Title,
		))
	}

	sb.WriteString("All other items in the context file are supporting context — ")
	sb.WriteString("related tickets, linked documentation, referenced pages — not the main subject.\n\n")

	sb.WriteString("---\n\n")
	sb.WriteString("## Instructions\n\n")
	sb.WriteString("1. Read the context file at the path above.\n")
	sb.WriteString("2. Explore the current working directory's codebase as needed to populate Investigation Notes.\n")
	sb.WriteString("3. Produce a complete plan file in Markdown, covering **all** sections defined in the spec above:\n")
	sb.WriteString("   - `# Goal`\n")
	sb.WriteString("   - `# Reference Material`\n")
	sb.WriteString("   - `# Need to Know`\n")
	sb.WriteString("   - `# Questions from AI Agent`\n")
	sb.WriteString("   - `# Investigation Notes`\n")
	sb.WriteString("   - `# Development Plan`\n")
	sb.WriteString("   - `# Test Plan`\n")
	sb.WriteString("   - `# Implementation Status`\n\n")
	sb.WriteString("4. Do **not** ask clarifying questions. Write the plan based on all available information. ")
	sb.WriteString("If there are genuine unknowns that require a human decision, record them as unanswered ")
	sb.WriteString("entries in `# Questions from AI Agent` rather than stopping to ask.\n\n")
	sb.WriteString("5. The Development Plan must be broken into phases. Each phase must contain a ")
	sb.WriteString("checklist of specific, file-level TODOs. Phases must leave the codebase in a valid ")
	sb.WriteString("state when complete.\n\n")
	sb.WriteString("6. Write the plan file to disk at the path you are given by the tool that invoked you. ")
	sb.WriteString("Do not print it to stdout — write it as a file.\n")

	return sb.String()
}
