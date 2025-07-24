# docs-template-update

A Go tool for updating package documentation templates to match the new Elastic standardized format.

## Features

- Checks if `_dev/build/docs/readme.md` exists
  - If no, creates one and copies content from the existing `docs/README.md`
  - If yes, uses the existing file
- Sends the content to Google Gemini AI to transform according to the new template
  - Replaces "Exported fields" sections with the mustache placeholder `{{fields "data_stream_name"}}`
  - Replaces "Sample events" sections with the mustache placeholder `{{event "data_stream_name"}}`
  - Syncs the document structure with the new template format
- Automatically detects data streams from the package directory structure
  - For single data stream packages: applies the data stream name to the placeholders
  - For multiple data stream packages: creates separate sections for each data stream
- Returns a git-patch that can be applied or modified before applying

## Requirements

- Go 1.24 or later
- A Google Gemini API key

## Installation

```bash
go install github.com/andrewkroh/go-examples/docs-template-update@latest
```

## Usage

```bash
# Basic usage with Google API key as environment variable
export GOOGLE_API_KEY="your-api-key"
docs-template-update -path /path/to/package

# Generate a patch without writing changes (dry run)
docs-template-update -path /path/to/package -dry-run > changes.patch

# Apply the generated patch
docs-template-update -path /path/to/package | git apply
```

### Command Line Options

```
  -api-key string
        Google Gemini API key (can also be set via GOOGLE_API_KEY environment variable)
  -dry-run
        Generate patch but don't write changes to file
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
4. The AI processes the content to match the new template format, with special handling for:
   - "Exported fields" sections → `{{fields "data_stream_name"}}`
   - "Sample events" sections → `{{event "data_stream_name"}}`
5. The tool automatically detects data stream names by examining the `data_stream` directory:
   - For a single data stream: replaces placeholders with the detected stream name
   - For multiple data streams: creates separate sections for each data stream with appropriate placeholders
6. The tool generates a git-compatible patch showing the changes
7. Unless in dry-run mode, it writes the updated content back to the file

## License

This project is licensed under the same terms as other projects in the go-examples repository.
