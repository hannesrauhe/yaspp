# Copilot Instructions for YASPP

## Repository Overview

YASPP (Yet Another Simple Podcast Publisher) is a simple podcast publishing system used for "Chaos im Radio" (Chaos Radio Potsdam). The system generates static RSS feeds and web players from YAML episode metadata.

## Architecture

This repository contains two main components:

### 1. Python Publishing System (`yaspp.py`)
- **Purpose**: Generates static podcast feeds and HTML web players from YAML content
- **Main script**: `yaspp.py` - CLI tool that processes `content.yaml` 
- **Templates**: `templates.py` - HTML templates for web player and RSS feed
- **Configuration**: `config.py` - podcast metadata, URLs, and theming
- **Dependencies**: Uses `podgen` for RSS generation, `pyyaml` for YAML parsing, `markdown` for content processing

### 2. Go Metadata Extractor (`pad2gh/`)
- **Purpose**: Extracts podcast episode metadata from HedgeDoc pad entries
- **Main file**: `pad2gh/main.go` - CLI tool that parses collaborative notes into structured YAML
- **Input**: HedgeDoc pad URLs (from `https://pad.ccc-p.org/`)
- **Output**: Appends episode data to `content.yaml` and creates PR comments in `comments.md`
- **Dependencies**: Uses `logrus` for logging, `gopkg.in/yaml.v3` for YAML generation

## Data Flow

1. Episode notes are collaboratively written in HedgeDoc pads
2. `pad2gh` tool extracts structured data from pad markdown sections:
   - `summary` → episode summary
   - `shownotes`/`long summary` → detailed show notes
   - `chapters` → timestamped chapter marks
   - `mukke` → music links and credits
3. Extracted data is appended to `content.yaml` in YAML frontmatter format
4. `yaspp.py` processes `content.yaml` to generate:
   - `feed.xml` - RSS podcast feed
   - `index.html` - Web player interface

## Code Patterns & Conventions

### Python Code (`yaspp.py`, `config.py`, `templates.py`, `util.py`)
- **Style**: Follow Python PEP 8 conventions
- **Dependencies**: Minimal external dependencies (see `requirements.txt`)
- **Templates**: Use Python `string.Template` for HTML generation
- **Configuration**: Centralized in `config.py` with podcast-specific settings
- **Error handling**: Use Python exceptions and proper error messages

### Go Code (`pad2gh/main.go`)
- **Style**: Follow Go standard formatting (`gofmt`)
- **Structs**: Use YAML tags for serialization (e.g., `yaml:"uuid"`)
- **Logging**: Use `logrus.StandardLogger()` for consistent logging
- **Error handling**: Return errors explicitly, use `logger.Fatalf()` for fatal errors
- **CLI**: Use `flag` package for command-line arguments

## Key Data Structures

### Episode YAML Format (in `content.yaml`)
```yaml
---
uuid: "nt-2024-01-15"           # Format: nt-YYYY-MM-DD
title: "CiR am 15.01.2024"     # Format: CiR am DD.MM.YYYY
subtitle: "Der Chaostreff im Freien Radio Potsdam"
summary: "Episode description..."
publicationDate: "2024-01-15T00:00:00+02:00"  # ISO 8601 format
audio:
  - url: "$media_base_url/2024_01_15-chaos-im-radio.mp3"
    mimeType: "audio/mp3"
chapters:
  - start: "00:00:00.000"
    title: "Chapter title"
long_summary_md: "**Shownotes:**\nMarkdown content..."
```

### Go Structs
- `CiREntry` - Main episode data structure
- `CiRaudio` - Audio file information  
- `CiRChapter` - Chapter/timestamp information

## Development Guidelines

### When Working on Python Code:
- Maintain compatibility with existing `config.py` settings
- Preserve HTML template structure in `templates.py`
- Use `podgen` library patterns for RSS feed generation
- Handle YAML parsing errors gracefully
- Test with sample `content.yaml` data

### When Working on Go Code:
- Parse HedgeDoc markdown by sections (headers starting with `##`)
- Extract URLs from markdown links using regex patterns
- Handle missing sections gracefully with informative error messages
- Use structured logging with appropriate log levels
- Validate date formats and URL patterns

### Common Tasks:
- **Adding new episode fields**: Update both Go structs and Python processing
- **Modifying templates**: Edit `templates.py` and test HTML output
- **Changing episode format**: Update `pad2gh/main.go` parsing logic
- **Adding new pad sections**: Extend `getMarkdownContentBySection()` function

## Testing & Development

### Running the System:
```bash
# Generate podcast files from YAML
python yaspp.py [-o output_dir] content.yaml

# Extract episode from HedgeDoc pad
cd pad2gh && go run main.go -l <pad_url> -o ../content.yaml -c ../comments.md
```

### Development Environment:
- Python 3.10+ with packages from `requirements.txt`
- Go 1.21+ with modules in `pad2gh/go.mod`
- Access to `https://pad.ccc-p.org/` for testing pad extraction

## File Organization

- `content.yaml` - Episode database (YAML frontmatter format)
- `yaspp.py` - Main publishing script
- `config.py` - Podcast configuration
- `templates.py` - HTML/RSS templates
- `pad2gh/main.go` - HedgeDoc pad parser
- `requirements.txt` - Python dependencies
- `util.py` - Chapter mark utilities
- `.github/workflows/` - GitHub Actions for automation

## Special Considerations

- **Media URLs**: Use `$media_base_url` placeholder, resolved in `config.py`
- **Date Formats**: Episode dates must be in `YYYY-MM-DD` format for parsing
- **Chapter Timing**: Use `HH:MM:SS.mmm` format for precise chapter marks
- **German Language**: Content is primarily in German (de-DE locale)
- **URL Validation**: HedgeDoc URLs must start with `https://pad.ccc-p.org/`
- **YAML Frontmatter**: Each episode in `content.yaml` is separated by `---`

## Automation

The repository includes GitHub Actions workflows:
- `create-pr-from-pad.yml` - Automatically extracts episodes from pads and creates PRs
- `docker-build-and-run.yml` - Builds and tests the publishing system

When modifying automation, ensure compatibility with the existing workflow structure and output formats.