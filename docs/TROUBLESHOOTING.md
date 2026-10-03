# Troubleshooting

## Gemini errors

HelAIx writes backend diagnostics to `helaix/helaix.log` below the platform user configuration directory:

- Windows: `%APPDATA%\helaix\helaix.log`
- macOS: `~/Library/Application Support/helaix/helaix.log`
- Linux: `~/.config/helaix/helaix.log`

The log records the operation, selected model, HTTP status, provider status, and provider error message. It does not record the API key, prompts, or generated preset content.

Log storage is capped at 1 MiB, individual entries are capped at 4 KiB, and the in-app viewer reads at most the most recent 1 MiB of the file.

For a `503`, reproduce the request once and inspect the matching `AI request failed` line. A `503` is a provider or temporary availability error, so also record the timestamp, model, and whether the connection test succeeds. Do not send the log if it contains personal information from a provider error message.

The connection test calls the Gemini `ListModels` endpoint. A successful connection test does not guarantee that a subsequent `GenerateContent` request will succeed: generation can have separate quota, model availability, or transient service failures.
