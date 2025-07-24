# docs-template-update

A Go tool for updating package documentation templates to match the new Elastic standardized format.

## Requirements

- Go 1.24 or later
- A Google Gemini API key

## Installation

```bash
go install github.com/kgeller/go-examples/docs-template-update@latest
```

## Usage

```bash
# Basic usage with Google API key as environment variable
export GOOGLE_API_KEY="your-api-key"
docs-template-update -path /path/to/package

# Apply the generated patch
docs-template-update -path /path/to/package | git apply
```

### How to create a Gemini API key

1. Go to the [Google AI Studio](https://makersuite.google.com/app/ai-studio)
2. Click on the "API keys" tab
3. Click on the "Create API key" button
4. Copy the API key

### Command Line Options

```
  -api-key string
        Google Gemini API key (can also be set via GOOGLE_API_KEY environment variable)
  -path string
        Path to the package directory (default ".")
  -verbose
        Enable verbose logging
```

## How It Works

1. The tool first checks if `_dev/build/docs/readme.md` exists in the specified package
   - If not, it creates the directory structure and copies the content from `docs/README.md`
2. It fetches the template from the Elastic Package repository
3. The tool sends both the existing content and template to the Google Gemini API
4. The AI processes the content to match the new template format and writes the updated content back to the file