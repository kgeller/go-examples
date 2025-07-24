package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/otiai10/copy"
	"github.com/pmezard/go-difflib/difflib"
	"google.golang.org/api/option"
	"google.golang.org/api/iterator"
)

const (
	templateURL = "https://raw.githubusercontent.com/elastic/elastic-package/89b34ec09f562b2c1c921ba4b465b6ef96ea47de/internal/packages/archetype/_static/package-docs-readme.md.tmpl"
	// System prompt for instructing the LLM
	systemPrompt = `You are a documentation expert specializing in Elastic documentation templates.
Your task is to transform the provided README file to conform to the new template structure. This is intended to be an additive process,
so do not remove any existing content, only restructure it to fit the new template.

Here is some context for you to reference for your task, read it carefully as you will get questions about it later:
# Original README content:
%s

# New template structure:
%s
`
	// User prompt template for the LLM
	userPromptTemplate = `I need to update this README.md file to match our new documentation template.

Follow these exact guidelines:
1. Always utilize the original content of the README.md file where possible
2. Restructure the document to follow the new template format provided
3. If any content is not relevant to the new template, copy it to the Reference section and add a note it in a code comment for why it should be removed
4. Do not include the following from the tempalte: initial comment from the template, the header placeholder, or the Reference -> ECS field reference section
5. Always organize the datastreams together under Reference section. For each datastream there should be
a brief summary, exported fields, and sample events sections all separated with an empty line.
6. Always prefix sample event placeholders with 'An example event for "data_stream_name" looks as following:'.
7. Format your response appropriately for a Markdown file
8. Replace any 'Exported fields' sections with the mustache placeholder: {{fields "data_stream_name"}}
9. Replace any 'Sample event' sections with the mustache placeholder: {{event "data_stream_name"}}
10. If there is no content for a section, you must add a code comment with some guidance to the user on what to add.
11. Sync the document with the new template structure

Return ONLY the updated Markdown content, without any explanation or commentary.`
)

var (
	googleAPIKey string
	packagePath  string
	verbose      bool
)

func init() {
	flag.StringVar(&googleAPIKey, "api-key", "", "Google Gemini API key (required)")
	flag.StringVar(&packagePath, "path", ".", "Path to the package directory")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "docs-template-update updates documentation templates to the new format.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if googleAPIKey == "" {
		googleAPIKey = os.Getenv("GOOGLE_API_KEY")
		if googleAPIKey == "" {
			log.Fatal("Google API key is required. Set it using the -api-key flag or GOOGLE_API_KEY environment variable")
		}
	}

	// Process the package
	patch, err := processPackage(packagePath)
	if err != nil {
		log.Fatalf("Error processing package: %v", err)
	}

	// Print the git patch
	fmt.Println(patch)
}

// findDataStreams discovers data stream directories in the package
func findDataStreams(pkgPath string) ([]string, error) {
	dataStreamPath := filepath.Join(pkgPath, "data_stream")
	
	// Check if data_stream directory exists
	if _, err := os.Stat(dataStreamPath); os.IsNotExist(err) {
		if verbose {
			log.Printf("No data_stream directory found at %s", dataStreamPath)
		}
		return nil, nil
	}
	
	// List directories in data_stream directory
	entries, err := os.ReadDir(dataStreamPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read data_stream directory: %w", err)
	}
	
	var dataStreams []string
	for _, entry := range entries {
		if entry.IsDir() {
			dataStreams = append(dataStreams, entry.Name())
		}
	}
	
	if verbose {
		log.Printf("Found data streams: %v", dataStreams)
	}
	
	return dataStreams, nil
}

// applyDataStreamPlaceholders replaces generic placeholders with specific data stream names
func applyDataStreamPlaceholders(content string, dataStreams []string) string {
	if len(dataStreams) == 0 {
		return content
	}

	// Create a regex pattern to find generic placeholders
	fieldsPattern := regexp.MustCompile(`\{\{fields\s+"data_stream_name"\}\}`)
	eventPattern := regexp.MustCompile(`\{\{event\s+"data_stream_name"\}\}`)
	
	// For each data stream, add a section with the proper placeholders
	var result strings.Builder
	
	// Check if there's a single data stream or multiple
	if len(dataStreams) == 1 {
		// If single data stream, just replace the placeholders
		result.WriteString(fieldsPattern.ReplaceAllString(content, fmt.Sprintf(`{{fields "%s"}}`, dataStreams[0])))
		content = result.String()
		result.Reset()
		result.WriteString(eventPattern.ReplaceAllString(content, fmt.Sprintf(`{{event "%s"}}`, dataStreams[0])))
		return result.String()
	}

	// For multiple data streams, we need more complex processing
	sections := strings.Split(content, "### ECS field Reference")
	if len(sections) != 2 {
		sections = strings.Split(content, "### Sample Event")
		if len(sections) != 2 {
			// If we can't find the headers, just replace with the first data stream
			if verbose {
				log.Println("Could not identify sections properly for multiple data streams, using first data stream")
			}
			result.WriteString(fieldsPattern.ReplaceAllString(content, fmt.Sprintf(`{{fields "%s"}}`, dataStreams[0])))
			content = result.String()
			result.Reset()
			result.WriteString(eventPattern.ReplaceAllString(content, fmt.Sprintf(`{{event "%s"}}`, dataStreams[0])))
			return result.String()
		}
	}

	// Handle multiple data streams by creating sections for each
	result.WriteString(sections[0])
	result.WriteString("### ECS field Reference\n\n")
	
	// Add fields sections for each data stream
	for _, ds := range dataStreams {
		result.WriteString(fmt.Sprintf("#### %s\n\n{{fields \"%s\"}}\n\n", ds, ds))
	}
	
	// If we can split by Sample Event header
	eventSections := strings.Split(sections[1], "### Sample Event")
	if len(eventSections) == 2 {
		result.WriteString("### Sample Event\n\n")
		
		// Add event sections for each data stream
		for _, ds := range dataStreams {
			result.WriteString(fmt.Sprintf("#### %s\n\n{{event \"%s\"}}\n\n", ds, ds))
		}
		
		result.WriteString(eventSections[1])
	} else {
		// Fallback if we can't find the Sample Event header
		result.WriteString(sections[1])
	}
	
	return result.String()
}

func processPackage(pkgPath string) (string, error) {
	// Ensure target directory exists
	targetDir := filepath.Join(pkgPath, "_dev", "build", "docs")
	targetPath := filepath.Join(targetDir, "readme.md")
	sourcePath := filepath.Join(pkgPath, "docs", "README.md")

	if verbose {
		log.Printf("Checking if target directory exists: %s", targetDir)
	}

	// Check if target readme exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		// Check if source readme exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			return "", fmt.Errorf("source README.md not found at %s", sourcePath)
		}

		// Copy the source readme to the target
		if verbose {
			log.Printf("Copying %s to %s", sourcePath, targetPath)
		}
		
		if err := copy.Copy(sourcePath, targetPath); err != nil {
			return "", fmt.Errorf("failed to copy README.md: %w", err)
		}
	}

	// Read the template from GitHub
	template, err := fetchTemplate()
	if err != nil {
		return "", fmt.Errorf("failed to fetch template: %w", err)
	}

	// Read the existing readme
	readmeContent, err := os.ReadFile(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to read readme: %w", err)
	}

	// Generate updated content using LLM
	updatedContent, err := generateUpdatedReadme(string(readmeContent), template)
	if err != nil {
		return "", fmt.Errorf("failed to generate updated readme: %w", err)
	}

	// Find data streams
	dataStreams, err := findDataStreams(pkgPath)
	if err != nil {
		return "", fmt.Errorf("failed to find data streams: %w", err)
	}
	
	// Apply data stream placeholders
	updatedContent = applyDataStreamPlaceholders(updatedContent, dataStreams)

	// Generate a diff/patch
	patch, err := generatePatch(targetPath, string(readmeContent), updatedContent)
	if err != nil {
		return "", fmt.Errorf("failed to generate patch: %w", err)
	}

	// Write the changes
	if err := os.WriteFile(targetPath, []byte(updatedContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write updated readme: %w", err)
	}
	if verbose {
		log.Printf("Updated readme written to %s", targetPath)
	}

	return patch, nil
}

func fetchTemplate() (string, error) {
	resp, err := http.Get(templateURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch template, status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func generateUpdatedReadme(readmeContent, templateContent string) (string, error) {
	// Create context with 5 minute timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	// Create a Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(googleAPIKey))
	if err != nil {
		return "", fmt.Errorf("error creating Gemini client: %w", err)
	}
	defer client.Close()

	// List available models for debugging if in verbose mode
	if verbose {
		log.Printf("Available models:")
		iter := client.ListModels(ctx)
		for {
			model, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("Error listing models: %v", err)
				break
			}
			log.Printf("- %s", model.Name)
		}
	}

	// Use the gemini-2.5-pro model directly
	modelName := "gemini-2.5-pro"
	if verbose {
		log.Printf("Using model: %s", modelName)
	}
	
	model := client.GenerativeModel(modelName)

	// Set safety settings to allow content generation
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockNone,
		},
	}

	// Build the complete prompt with system instructions and user content
	completePrompt := fmt.Sprintf("%s\n\n%s", fmt.Sprintf(systemPrompt, readmeContent, templateContent), userPromptTemplate)	
	// Send the request
	resp, err := model.GenerateContent(ctx, genai.Text(completePrompt))
	if err != nil {
		return "", fmt.Errorf("error generating content with %s: %w", modelName, err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response received from Gemini")
	}

	// Extract the text response
	responseText, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response type from Gemini")
	}

	return string(responseText), nil
}

func generatePatch(filePath, original, updated string) (string, error) {
	fromLines := strings.Split(original, "\n")
	toLines := strings.Split(updated, "\n")

	diff := difflib.UnifiedDiff{
		A:        fromLines,
		B:        toLines,
		FromFile: "a/" + filepath.Base(filePath),
		ToFile:   "b/" + filepath.Base(filePath),
		Context:  3,
	}

	return difflib.GetUnifiedDiffString(diff)
}
