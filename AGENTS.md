# HelAIx Agent Guidelines

Welcome, AI Agent! This file contains instructions and rules you MUST follow when working on the HelAIx project.

## Architecture Context
- **Core Technology:** This is a Wails desktop application. It uses a Go backend (`app/main.go`, `app/pkg/`) and a React frontend (`app/frontend/`).
- **Goal:** HelAIx is an AI-powered Line 6 Helix preset generator. It converts natural language into `.hlx` JSON preset files using the Gemini API.
- **Native OS Focus:** The app is compiled into a single native executable. We aim to keep binaries small and performant. **The application must remain compatible with all supported operating systems: Windows, MacOS, and Linux (best effort).** Avoid introducing dependencies or shell commands that only work on one platform.

## Coding Standards (For AI and Humans)

### General Rules
- Always refer to `docs/STYLEGUIDE.md` for detailed coding standards. 
- Ensure you respect the architectural separation between the Go backend and React frontend.
- Never hardcode credentials, API keys, or sensitive data.

### Go Backend (`app/`, `app/pkg/`)
- Adhere to standard Go formatting.
- Use the `wails` context properly for emitting events to the frontend.
- Structure packages logically (e.g., `pkg/helix`, `pkg/gemini`, `pkg/config`).
- Error handling must be explicit and informative. Do not silently swallow errors.

### React Frontend (`app/frontend/`)
- Write functional components with React Hooks.
- Use Wails JS bindings (generated in `app/frontend/wailsjs/`) to communicate with the Go backend. Do not create raw HTTP or WebSocket connections for internal app communication.

## AI-Specific Constraints
- **Do not modify the `wails.json` or core build scripts unless explicitly requested.**
- Before adding new heavy dependencies (npm or go mod), confirm with the user.
- When generating or validating Helix presets, ensure the generated JSON strictly adheres to the `.hlx` format rules documented in the project.
