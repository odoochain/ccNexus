# ccNexus Architecture and Development Guide

## Project Overview

ccNexus (Claude Code Nexus) is a smart API endpoint rotation proxy designed for Claude Code. It provides automatic failover, load balancing, and comprehensive statistics for managing multiple API endpoints, with support for both official Claude API and third-party providers.Always respond in Chinese-simplified.

### Key Characteristics

- Language: Go (backend ~1,280 lines) + JavaScript (frontend)
- Framework: Wails v2 (Go + Web framework for desktop)
- Build System: Wails CLI with Vite for frontend
- Platforms: Windows, macOS (Intel/ARM64), Linux
- Dependencies: Standard library + Wails only

## High-Level Architecture

### System Design - Three Tiers

The application uses a three-tier architecture:

1. **Frontend Layer**: JavaScript/HTML/CSS running in Wails WebView
   - User interface for endpoint management
   - Real-time statistics dashboard
   - Log viewer with filtering

2. **Wails Bridge Layer**: Go-to-JavaScript communication
   - Auto-generated method bindings
   - Application lifecycle management
   - Configuration persistence

3. **Backend Services**: Go microservices
   - HTTP Proxy: Forwards requests with endpoint rotation
   - Config Manager: Persistent JSON-based configuration
   - Logger: Multi-level circular-buffer logging
   - Statistics: Real-time per-endpoint metrics

### Request Processing Flow

Claude Code → http://localhost:3000 → Proxy Server →
  1. Select enabled endpoint (round-robin)
  2. Forward request with endpoint API key
  3. Parse response (extract tokens)
  4. On non-200: rotate endpoint and retry
  5. On 200: return response to Claude Code
  6. Record statistics and log activity

## Backend Architecture (Go)

### File-by-File Component Breakdown

**main.go (66 lines)**: Application Entry Point
- Initializes Wails application
- Embeds frontend assets using go:embed
- Configures window size (1024x768)
- Sets platform-specific options
- Handles application initialization

**app.go (319 lines)**: Wails Bridge Layer
- Bridges JavaScript frontend to Go backend
- Manages application lifecycle (startup/shutdown)
- All methods automatically exposed to frontend
- Handles configuration CRUD operations
- Manages endpoint operations
- Provides logging and statistics access

Key Exposed Methods:
- Configuration: GetConfig(), UpdateConfig(), UpdatePort()
- Endpoints: AddEndpoint(), RemoveEndpoint(), UpdateEndpoint(), ToggleEndpoint()
- Statistics: GetStats()
- Logging: GetLogs(), SetLogLevel(), ClearLogs()

**internal/proxy/proxy.go (457 lines)**: HTTP Proxy Engine
- Main HTTP proxy server listening on :3000
- handleProxy(): Main request handler with retry logic
- rotateEndpoint(): Round-robin endpoint selection
- Handles both streaming (SSE) and JSON responses
- Extracts token usage from API responses
- 120-second request timeout
- Gzip decompression support

Request Handling Algorithm:
1. Read request body into buffer
2. Get list of enabled endpoints
3. For each retry (max = endpoint count):
   a. Select current endpoint (round-robin)
   b. Create forwarding request with endpoint API key
   c. Send request with 120-second timeout
   d. On non-200: extract error, rotate, retry
   e. On 200: extract tokens, return response

**internal/proxy/stats.go (100 lines)**: Statistics Engine
- Thread-safe metrics tracking per endpoint
- Metrics: requests, errors, input/output tokens, last used
- Token extraction from SSE and JSON
- Cache token support (cache_read_input_tokens, etc)
- RecordRequest(), RecordError(), RecordTokens() methods
- GetStats() returns deep copy (prevents mutation)

**internal/config/config.go (163 lines)**: Configuration Manager
- Thread-safe config with getter/setter methods
- Persistent JSON storage at ~/.ccNexus/config.json
- Configuration validation on load and save
- Auto-creates directory if missing
- Graceful defaults on file not found
- Immediate save on any change
- Thread-safe access with sync.RWMutex

Configuration Schema:
{
  "port": 3000,
  "logLevel": 1,
  "endpoints": [ { "name": "...", "apiUrl": "...", "apiKey": "...", "enabled": true } ]
}

**internal/logger/logger.go (177 lines)**: Logging System
- Singleton logger with 4 levels: DEBUG(0), INFO(1), WARN(2), ERROR(3)
- Circular buffer (max 1000 entries)
- Level-based filtering (minimum threshold)
- Both console output and in-memory storage
- Thread-safe with mutex protection
- Emoji icons for visual identification
- Auto-formats timestamps and messages

## Frontend Architecture (JavaScript)

### Main Components

**frontend/src/main.js (1000+ lines)**: Single-Page UI Application
- Provides complete user interface
- Real-time statistics dashboard
- Endpoint management (add/edit/remove/toggle)
- Log viewer with level filtering
- Modal dialogs for data entry
- Wails integration via window.go.main.App.* methods

UI Components:
1. Header: Logo, title, port display, GitHub link
2. Statistics: Endpoints count, request stats, token usage
3. Endpoints Panel: List with add/edit/remove/toggle
4. Logs Panel: Real-time viewer, level filter, copy/clear
5. Modals: Add/edit endpoint, port config, welcome

Auto-Refresh:
- Statistics: Every 5 seconds
- Logs: Every 2 seconds

**frontend/src/style.css**: Responsive Design
- Card-based component layout
- Grid system for statistics
- Dark theme for logs panel
- Smooth animations and transitions
- Responsive buttons and modals

## Communication Layer (Wails)

### Frontend-Backend Bridge

Wails auto-generates TypeScript/Go bindings for all methods.
Frontend calls are asynchronous (Promise-based).

Method Categories:

Configuration Methods:
- window.go.main.App.GetConfig() - Get current config
- window.go.main.App.UpdateConfig(json) - Update config
- window.go.main.App.UpdatePort(port) - Change port

Endpoint Methods:
- window.go.main.App.AddEndpoint(name, url, key)
- window.go.main.App.RemoveEndpoint(index)
- window.go.main.App.UpdateEndpoint(index, name, url, key)
- window.go.main.App.ToggleEndpoint(index, enabled)

Statistics Methods:
- window.go.main.App.GetStats() - Get all stats

Logging Methods:
- window.go.main.App.GetLogs() - Get all logs
- window.go.main.App.GetLogsByLevel(level) - Filtered logs
- window.go.main.App.SetLogLevel(level) - Set minimum level
- window.go.main.App.ClearLogs() - Clear buffer
- window.go.main.App.GetLogLevel() - Current level

## Build System and Development

### Prerequisites

- Go 1.22+
- Node.js 18+
- Wails CLI v2: go install github.com/wailsapp/wails/v2/cmd/wails@latest

### Development Workflow

Setup:
git clone https://github.com/lich0821/ccNexus.git
cd ccNexus
go mod download
cd frontend && npm install -g npm@latest
 %% npm install && cd ..

Development:
wails dev
- Hot reload for JS/CSS/HTML
- Auto-recompile for Go
- Frontend dev server: http://localhost:34115

Build:
wails build                              # Current platform
wails build -platform windows/amd64     # Specific platform
wails build -platform darwin/amd64      # macOS Intel
wails build -platform darwin/arm64      # macOS Apple Silicon
wails build -platform linux/amd64       # Linux
Output: build/bin/ccNexus

### CI/CD Pipeline

GitHub Actions (.github/workflows/build.yml):
- Triggered on version tags (v*) or manual dispatch
- Multi-platform builds in parallel:
  * Ubuntu → Linux amd64 (tar.gz)
  * macOS → macOS amd64 (zip)
  * macOS → macOS arm64 (zip)
  * Windows → Windows amd64 (zip)
- Creates GitHub Release with artifacts
- Auto-generates release notes

## Key Architectural Decisions

### 1. Round-Robin Load Balancing Strategy
- Simple modulo-based rotation through endpoints
- Only enabled endpoints participate in rotation
- On error: immediately rotate to next endpoint
- On success: keep same endpoint for cache efficiency
- Ensures fair distribution and failover capability

### 2. Retry on Any Non-200 Status
- Retries on ALL non-200 responses (429, 500, 503, etc)
- Not just connection errors
- Ensures requests succeed if any endpoint is available
- Max retries = number of enabled endpoints
- Provides robust failover mechanism

### 3. Immediate Token Extraction During Request
- Tokens extracted during response processing
- Not post-processed or batched
- Handles both SSE streaming and JSON formats
- Supports cache token variants
- Real-time statistics recording

### 4. Mutex-Based Thread Synchronization
- Read-write locks for config (frequently accessed)
- Write locks for statistics updates
- Simple, straightforward synchronization model
- No lock-free or complex concurrent patterns
- Clear ownership and lifecycle

### 5. Persistent Configuration Storage
- Immediate save on any change (not batched)
- JSON file format (human-readable)
- Validation before save
- Graceful defaults for missing values
- Local-first approach (no cloud sync)

### 6. Singleton Logger with Circular Buffer
- One logger instance per application
- Circular buffer prevents unbounded memory growth
- Filtering at log-time (not query-time)
- Both console and in-memory UI display
- Level-based minimum threshold

### 7. Polling-Based Frontend Updates
- No WebSocket complexity
- Simple request-response pattern
- Stats polled every 5 seconds
- Logs polled every 2 seconds
- Wails auto-generates method bindings

## Important Implementation Details

### Endpoint Rotation Mechanism
- Maintains currentIndex in Proxy struct
- Incremented only on error (not on success)
- Uses modulo to wrap around endpoints
- Only enabled endpoints considered
- Logged at INFO level when rotation occurs

### Token Extraction from Streaming (SSE)
- Parses Server-Sent Events line-by-line
- Extracts from message_start event:
  * usage.input_tokens
  * usage.cache_read_input_tokens
  * usage.cache_creation_input_tokens
  * usage.output_tokens (initial)
- Also checks message_delta event for final output count
- Handles multiple event types within single response

### Token Extraction from JSON
- Standard JSON unmarshaling
- Reads response.usage.input_tokens and output_tokens
- Handles partial data gracefully
- Falls back if JSON format unexpected

### Gzip Decompression
- Detects gzip by magic bytes: 0x1f 0x8b
- Decompresses for token extraction
- Returns original compressed response to client
- Transparent to client application

### HTTP Header Manipulation
- x-api-key: Overwritten with endpoint API key
- Host: Overwritten with endpoint domain
- All other headers: Passed through unchanged
- Original request headers processed before forwarding

### Statistics Persistence
- Not cleared on configuration changes
- Not cleared when toggling endpoints
- Only cleared by explicit Clear() call
- Persists for entire application lifetime
- Resets on application restart (not persisted to disk)

## File Organization

Project Structure:

ccNexus/
├── main.go                  # Application entry (66 lines)
├── app.go                   # Wails bridge (319 lines)
├── go.mod, go.sum          # Go dependencies
├── wails.json              # Wails configuration
├── internal/
│   ├── proxy/
│   │   ├── proxy.go        # HTTP proxy engine (457 lines)
│   │   └── stats.go        # Statistics tracking (100 lines)
│   ├── config/
│   │   └── config.go       # Config management (163 lines)
│   └── logger/
│       └── logger.go       # Logging system (177 lines)
├── frontend/
│   ├── src/
│   │   ├── main.js         # UI implementation (1000+ lines)
│   │   └── style.css       # Styling
│   ├── index.html          # HTML entry point
│   ├── vite.config.js      # Vite configuration
│   ├── package.json        # npm dependencies
│   ├── public/             # Static assets
│   ├── dist/               # Built output
│   └── wailsjs/            # Auto-generated bindings
├── build/
│   ├── appicon.png/svg    # Application icons
│   └── bin/               # Build output
├── .github/workflows/      # CI/CD
├── README.md, README_CN.md # Documentation
├── LICENSE                 # MIT License
└── .gitignore              # Git ignore

## Configuration Reference

### Storage Location
- Windows: %USERPROFILE%\.ccNexus\config.json
- macOS/Linux: ~/.ccNexus/config.json

### Default Configuration
{
  "port": 3000,
  "logLevel": 1,
  "endpoints": [
    {
      "name": "Claude Official",
      "apiUrl": "api.anthropic.com",
      "apiKey": "your-api-key-here",
      "enabled": true
    }
  ]
}

### Log Levels
- 0 = DEBUG: Detailed debugging information
- 1 = INFO: General information (default)
- 2 = WARN: Warning messages
- 3 = ERROR: Error messages only

## Performance Characteristics

### Memory Usage
- Logging buffer: ~1MB (1000 entries at ~1KB each)
- Configuration: <1KB
- Statistics: ~100 bytes per endpoint
- Frontend (WebView): 30-50MB
- Total baseline: 35-55MB

### CPU Usage
- Idle: <1% (polling-based)
- Per request: 5-10ms proxy overhead
- Token extraction: 1-5ms inline

### Request Latency
- Proxy overhead: 5-10ms
- No response caching (pass-through)
- Request timeout: 120 seconds
- Network latency dominates total time

## Integration with Claude Code

### Configuration in Claude Code
- API Base URL: http://localhost:3000
- API Key: Any value (will be replaced by proxy)

### How Integration Works
1. Claude Code sends requests to localhost:3000
2. Proxy intercepts the request
3. Injects endpoint-specific API key via x-api-key header
4. Forwards to configured endpoint
5. Returns response unchanged to Claude Code
6. If endpoint fails: rotates and retries automatically

### Endpoint Compatibility
- Must support Anthropic Claude API format
- HTTPS protocol required
- Standard JSON request/response structure
- Can be official Anthropic API or third-party provider

## Unique Architectural Aspects

1. **Purpose-Built Proxy**: Designed specifically for Claude Code integration
2. **Simple & Reliable**: Basic round-robin algorithm, consistent behavior
3. **Real-Time Monitoring**: Live logs and statistics dashboard
4. **Local-First Privacy**: All data stored locally, no cloud synchronization
5. **Zero External Dependencies**: Backend uses only stdlib + Wails
6. **Single Binary Distribution**: Embedded frontend assets, self-contained
7. **Cross-Platform**: Native builds for Windows, macOS, Linux
8. **Efficient**: Minimal overhead (5-10ms per request)
9. **Maintainable**: Clear separation of concerns, modular design
10. **User-Friendly**: Beautiful, intuitive desktop interface

## Development Patterns and Tips

### Debugging Proxy Behavior
1. Set log level to DEBUG via UI or config file
2. Open Logs panel in application
3. Look for [ENDPOINT_NAME] prefixed messages
4. Token usage logged at DEBUG level
5. Endpoint rotations logged at INFO level
6. Errors logged at WARN level

### Testing Endpoints
Check health: curl http://localhost:3000/health
View statistics: curl http://localhost:3000/stats
Test proxy: curl -X POST http://localhost:3000/messages -H "..." -d "{...}"

### Adding New Features
1. Implement functionality in appropriate Go package
2. Add method to App struct if exposing to frontend
3. Wails automatically generates bindings
4. Call from JavaScript via window.go.main.App.METHOD()
5. Update frontend UI accordingly

### Understanding Statistics Flow
1. handleProxy() in proxy.go records requests
2. stats.RecordRequest() increments counters
3. On success: stats.RecordTokens() adds token usage
4. app.GetStats() retrieves and returns as JSON
5. Frontend polls every 5 seconds and updates UI

## Known Limitations

- Statistics reset on application restart (not persisted)
- No request caching between endpoints
- Port changes require application restart
- Single instance only (no clustering)
- Polling-based updates (not server push)
- No built-in request rate limiting

## Future Enhancement Possibilities

- WebSocket-based real-time updates
- Endpoint health check monitoring
- Request rate limiting per endpoint
- Custom header injection rules
- Advanced routing rules
- Prometheus metrics export
- Persistent statistics database
- Automatic endpoint selection algorithm
- Request caching between endpoints

## Quick Reference

### Commands
wails dev                              # Development with hot reload
wails build                            # Build for current platform
wails build -platform windows/amd64    # Build for specific platform

### Default Values
- Proxy Port: 3000
- Log Level: INFO (1)
- Config Path: ~/.ccNexus/config.json
- Window Size: 1024x768

### Environment Requirements
- Go 1.22 or later (1.24.3 recommended)
- Node.js 18 or later
- Wails v2.10.2
- Platform-specific build tools (see Wails documentation)

## Summary

ccNexus is a purpose-built reverse proxy for Claude Code that demonstrates clean architecture, efficient performance, and user-friendly interface. The three-tier design (frontend/bridge/backend) provides clear separation of concerns while maintaining simplicity. Automatic endpoint rotation with intelligent retry logic ensures reliable operation, while real-time monitoring provides visibility into proxy behavior. The codebase balances maintainability with functionality, making it suitable for both production use and educational purposes.

