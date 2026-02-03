# Building HelAIx for Linux

This guide provides instructions for building HelAIx on Linux. HelAIx is built using the [Wails](https://wails.io/) framework, which combines Go and React.
The application runs well on Linux, but you will have to either use a VM to run HX Edit or switch to Windows or MacOS to upload the files to your device.

## Prerequisites

Before building HelAIx, ensure you have the following installed:

### 1. System Dependencies (Standard for Ubuntu/Zorin OS)
Wails requires several system libraries for the GUI and webview. Run the following command to install them:

```bash
sudo apt update
sudo apt install build-essential libgtk-3-dev libwebkit2gtk-4.1-dev pkg-config
```

> [!NOTE]
> For Zorin OS 18 (based on Ubuntu 24.04), we recommend `libwebkit2gtk-4.1-dev`. If you are on an older version, you might need `libwebkit2gtk-4.0-dev`.

### 2. Go (Golang)
HelAIx requires Go 1.21 or later.
- Download it from [go.dev](https://go.dev/dl/).
- Ensure your `GOPATH` and `GOBIN` are set correctly, and `~/go/bin` is in your `PATH`.

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### 3. Node.js & NPM
Requires Node.js 15+.
- We recommend using [nvm](https://github.com/nvm-sh/nvm) or [mise](https://mise.jdx.dev/) to manage Node versions.

### 4. Wails CLI
Install the Wails CLI by running:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## Build Instructions

1. **Clone the repository**:
   ```bash
   git clone https://github.com/MrCitron/helAIx.git
   cd helAIx
   ```

2. **Install Frontend Dependencies**:
   ```bash
   cd app/frontend
   npm install
   cd ..
   ```

3. **Check Environment**:
   Run the following command to ensure all dependencies are correctly installed:
   ```bash
   wails doctor
   ```

4. **Build the Application**:
   To build a production binary:
   ```bash
   wails build
   ```

   > [!TIP]
   > On newer systems (like Zorin OS 18), if you encounter webview issues or if `wails doctor` reports a missing `libwebkit` dependency, build with the `webkit2_41` tag:
   > ```bash
   > wails build -tags webkit2_41
   > ```

The executable will be generated in the `app/build/bin` directory.

## Troubleshooting

### `wails doctor` reports `libwebkit` Not Found
On newer Linux distributions (Ubuntu 24.04 and its derivatives like Zorin OS 18), the `wails doctor` command might report that `libwebkit` is missing even if you have `libwebkit2gtk-4.1-dev` installed. This is because `doctor` specifically looks for the older `4.0` version.

**This can be safely ignored** as long as you have the `4.1` version installed and use the build tag `-tags webkit2_41`.

To verify you have the correct library installed, you can run:
```bash
pkg-config --modversion webkit2gtk-4.1
```
If it returns a version number (e.g., `2.4x.x`), you are ready to build.

## Development Mode

To run HelAIx in development mode with hot-reloading:
```bash
cd app
wails dev
```
