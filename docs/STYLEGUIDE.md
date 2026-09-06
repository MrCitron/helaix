# HelAIx Style Guide

This document outlines the coding standards for human developers and AI agents working on HelAIx.

## 1. Go Backend (Wails)

### Formatting & Quality
- All Go code must be formatted using standard Go tools (`gofmt`).
- Keep functions focused and concise.

### Error Handling
- Errors should be propagated and handled gracefully.
- When passing errors to the frontend via Wails, ensure they are clean, user-friendly strings, while logging the technical details on the backend.

### Wails Integration
- Only expose necessary Go methods to the frontend via the Wails bridge (`app.go`).
- Keep the core business logic isolated in `app/pkg/` packages (e.g., Gemini interaction, Helix file manipulation) rather than cluttering `app.go`.

## 2. React Frontend

### Component Structure
- Use Functional Components and React Hooks exclusively.
- Prefer keeping one main component per file.
- File naming convention: PascalCase for components (e.g., `PresetGenerator.jsx`), camelCase for utility scripts.

### State and Data Flow
- Use standard Hooks (`useState`, `useEffect`) for component state.
- For backend communication, strictly use the generated Wails bindings found in `wailsjs/go/`.

### Styling & UI
- Maintain a clean, native-feeling desktop interface.
- Ensure the layout is responsive and handles window resizing gracefully.

## 3. General Best Practices
- **Cross-Platform Compatibility:** The application must remain compatible with the different supported OS (Windows, MacOS, and Linux on a best-effort basis). Avoid platform-specific code or shell execution where possible. If strictly necessary, use Go's build constraints (build tags).
- **Commit Messages:** Use conventional commits (e.g., `feat:`, `fix:`, `docs:`, `refactor:`).
- **Comments:** Comment the "why" and not the "what". If the code implements complex logic (like parsing the proprietary Helix JSON format), document the assumptions.
- **Testing:** Write unit tests for core logic, especially for the `.hlx` file generation and Gemini API parsing.
