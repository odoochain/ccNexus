# ccNexus Architecture and Development Guide

## Project Overview

**ccNexus** (Claude Code Nexus) is a smart API endpoint rotation proxy designed for Claude Code. It provides automatic failover, load balancing, and real-time statistics for managing multiple API endpoints.

### Key Statistics
- Language: Go (backend) + JavaScript (frontend)
- Framework: Wails v2
- Backend Code: ~1,280 lines
- Platforms: Windows, macOS, Linux

## Architecture Overview

### Components
1. Frontend: JavaScript/HTML/CSS UI (Vite)
2. Bridge: Wails app connecting JS to Go
3. Backend: HTTP proxy, config, logging, stats
