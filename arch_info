# go-gaussian: Linear Equation Solver

A desktop GUI application for solving systems of 3x3 linear equations using Gaussian elimination, built with Go and the Ebiten game engine.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Core Components](#core-components)
- [Mathematical Engine](#mathematical-engine)
- [User Interface System](#user-interface-system)
- [Dependencies](#dependencies)
- [Setup & Installation](#setup--installation)
- [Usage](#usage)
- [File Structure](#file-structure)
- [Technical Specifications](#technical-specifications)

## Overview

go-gaussian implements an educational desktop application that demonstrates the Gaussian elimination algorithm through interactive visualization. The application emphasizes educational value and visual presentation over distributed architecture or high-performance computing.

### Key Features

- **Interactive Equation Input**: Three equation input fields with real-time validation
- **Animated Solution Process**: Step-by-step visualization of Gaussian elimination
- **Matrix Operations Visualization**: Live display of matrix transformations
- **File Output Generation**: Automatic solution export to timestamped files
- **Dynamic Window Management**: Responsive layout with content-based sizing
- **Comprehensive Error Handling**: User-friendly validation and error messages

## Architecture

The application follows a **game-loop architecture pattern** implemented through the Ebiten framework, with all core logic centralized in the `Game` struct. This monolithic design prioritizes simplicity and educational clarity.

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
├─────────────────┬─────────────────┬─────────────────────────┤
│   main()        │   NewGame()     │   Game struct           │
│   (lines 590-   │   (lines 562-   │   (lines 48-67)         │
│   603)          │   588)          │                         │
└─────────────────┼─────────────────┼─────────────────────────┘
                  │                 │
┌─────────────────┴─────────────────┴─────────────────────────┐
│                    UI System                                │
├─────────────────┬─────────────────┬─────────────────────────┤
│   Update()      │   Draw()        │   Layout()              │
│   (lines 284-   │   (lines 386-   │   (lines 265-282)       │
│   341)          │   459)          │                         │
└─────────────────┼─────────────────┼─────────────────────────┘
                  │                 │
┌─────────────────┴─────────────────┴─────────────────────────┐
│                Mathematical Core                            │
├─────────────────┬─────────────────┬─────────────────────────┤
│   Matrix        │   parseEquation │   GaussianElimination   │
│   struct        │   ()            │   ()                    │
│   (lines 42-46) │   (lines 177-   │   (lines 107-176)       │
│                 │   234)          │                         │
└─────────────────┴─────────────────┴─────────────────────────┘
```

## Core Components

### Game Struct (Central Controller)

The `Game` struct serves as the central application state manager, containing all necessary fields for equation input, solution processing, and UI state management.

```go
type Game struct {
    // UI State Management
    equations           []string
    activeEquation      int
    solving             bool
    solutionComplete    bool
    ShowExitPrompt      bool
    
    // Mathematical Processing
    matrix              *Matrix
    solution            string
    steps               []string
    currentStep         int
    
    // Rendering & Display
    font                font.Face
    closeButton         Button
    width               int
    height              int
    errorMsg            string
    solutionDisplayDone bool
}
```

### Matrix Struct (Mathematical Engine)

The `Matrix` struct encapsulates all linear algebra operations with specialized methods for Gaussian elimination.

```go
type Matrix struct {
    data [][]float64
    rows int
    cols int
}

// Core Operations
func (m *Matrix) SwapRows(i, j int)
func (m *Matrix) MultiplyRow(row int, scalar float64)
func (m *Matrix) AddMultipleOfRow(target, source int, scalar float64)
func (m *Matrix) GaussianElimination() ([]float64, []string, error)
```

### Button Struct (UI Elements)

Interactive UI components with click detection capabilities.

```go
type Button struct {
    X, Y, Width, Height int
    Text                string
}

func (b Button) Contains(x, y int) bool
```

## Mathematical Engine

The mathematical engine is organized into four distinct processing layers:

### 1. Input Processing Layer

**Function**: `parseEquation()` (lines 177-234)

- Converts string equations to coefficient arrays
- Handles equation normalization and tokenization
- Uses regex pattern `[+-]?\\d*\\.?\\d*[xyz]|[+-]?\\d+\\.?\\d*` for parsing
- Validates equation format and variable consistency

**Example Transformation**:
```
"2X + Y - Z = 8" → "2x+y-z=8" → [2.0, 1.0, -1.0, 8.0]
```

### 2. Matrix Operations Layer

**Core Methods**:
- `SwapRows(i, j int)` - Row interchange operations (lines 82-84)
- `MultiplyRow(row int, scalar float64)` - Scalar multiplication (lines 86-90)
- `AddMultipleOfRow(target, source int, scalar float64)` - Row addition (lines 92-96)

### 3. Algorithm Core

**Function**: `GaussianElimination()` (lines 107-176)

- Implements forward elimination and back substitution
- Precision control with `isZero()` and `round()` functions
- Records each step for educational visualization
- Handles edge cases (no solution, infinite solutions)

**Algorithm Flow**:
```
1. Forward Elimination
   ├── Find pivot element
   ├── Swap rows if necessary
   ├── Eliminate column entries
   └── Record transformation step
   
2. Back Substitution
   ├── Start from last row
   ├── Calculate variable values
   ├── Substitute into previous equations
   └── Record solution steps
```

### 4. Solution Generation

**Function**: `solve()` (lines 461-547)

- Formats solution output for display
- Generates timestamped solution files
- Creates step-by-step solution documentation
- Handles file I/O to `solutions/` directory

## User Interface System

### Game Loop Implementation

The UI follows Ebiten's game loop pattern with three core methods:

#### Update() Method (lines 284-341)
- Handles user input events
- Manages animation timing
- Updates application state
- Processes equation validation

#### Draw() Method (lines 386-459)
- Renders all visual elements
- Manages text positioning and formatting
- Displays matrix transformations
- Shows step-by-step solutions

#### Layout() Method (lines 265-282)
- Manages window dimensions
- Handles responsive scaling
- Returns current screen layout

### Input Handling

**Function**: `handleInput()` (lines 343-385)

- Keyboard input processing for equation entry
- Mouse click detection for UI interactions
- Equation field navigation (Tab, Enter, Arrow keys)
- Button interaction handling

### Animation System

The application implements a step-by-step animation system:

- **Display Timing**: 1800ms (30 seconds) per step
- **Step Management**: `currentStep` field tracks animation progress
- **Visual Feedback**: Real-time matrix transformation display
- **User Control**: Animation can be paused/resumed

## Dependencies

### Direct Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/hajimehoshi/ebiten/v2` | v2.8.6 | Game engine and GUI framework |
| `golang.org/x/image` | v0.24.0 | Extended image processing for font rendering |

### Indirect Dependencies (Grouped by Purpose)

**Platform Integration**:
- `github.com/ebitengine/gomobile`
- `github.com/ebitengine/hideconsole`
- `github.com/ebitengine/purego`

**System Operations**:
- `golang.org/x/sys`

**Text Processing**:
- `golang.org/x/text`

### Dependency Architecture

The mathematical engine has **minimal external dependencies** and could be extracted as a standalone package. Only the UI layer requires Ebiten framework dependencies.

```
Mathematical Core (Pure Go)
├── No external dependencies
├── Portable to other UI frameworks
└── Could be CLI/web compatible

UI Layer (Ebiten-dependent)
├── Ebiten game engine
├── Font rendering libraries
└── Platform-specific integrations
```

## Setup & Installation

### Prerequisites

- Go 1.16 or higher
- Git

### Installation

```bash
# Clone repository
git clone https://github.com/saedarm/go-gaussian.git
cd go-gaussian

# Download dependencies
go mod download

# Build application
go build -o go-gaussian main.go

# Run application
./go-gaussian
```

### Development Setup

```bash
# Install development dependencies
go mod tidy

# Run in development mode
go run main.go

# Run tests (if available)
go test ./...
```

## Usage

1. **Launch Application**: Run the executable
2. **Enter Equations**: Click on equation fields and input linear equations
   - Format: `aX + bY + cZ = d`
   - Example: `2X + 3Y - Z = 7`
3. **Solve System**: Click solve button to begin Gaussian elimination
4. **Watch Animation**: Observe step-by-step matrix transformations
5. **View Results**: Solution is displayed and saved to `solutions/` directory

### Input Format

**Supported Formats**:
- Standard form: `2X + 3Y - Z = 7`
- Mixed case: `2x + 3Y - z = 7`
- Decimal coefficients: `2.5X + 1.3Y - 0.7Z = 4.2`
- Negative coefficients: `-X + 2Y + 3Z = -1`

## File Structure

```
go-gaussian/
├── main.go                 # Main application file (604 lines)
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── solutions/              # Generated solution files
│   └── solution_*.txt      # Timestamped solutions
└── README.md              # Project documentation
```

### Code Organization (main.go)

| Lines | Component | Description |
|-------|-----------|-------------|
| 1-22 | Imports | Package imports and dependencies |
| 24-31 | Constants | Application constants and configuration |
| 33-67 | Structs | Data structure definitions |
| 69-106 | Utility Functions | Helper functions and validation |
| 107-176 | Mathematical Core | Gaussian elimination implementation |
| 177-234 | Input Processing | Equation parsing and validation |
| 235-282 | Layout Management | Window and UI layout handling |
| 284-341 | Update Logic | Game loop update method |
| 343-385 | Input Handling | User input processing |
| 386-459 | Rendering | Draw method and visual rendering |
| 461-547 | Solution Processing | Solution generation and file output |
| 549-588 | Initialization | Font loading and game setup |
| 590-603 | Main Function | Application entry point |

## Technical Specifications

### System Constants

```go
const (
    screenWidth  = 800    // Default window width
    screenHeight = 600    // Default window height
    fontSize     = 20     // UI text font size
    displayTime  = 1800   // Animation step duration (ms)
    minWidth     = 800    // Minimum window width
    minHeight    = 600    // Minimum window height
)
```

### Performance Characteristics

- **Memory Usage**: Minimal (single-threaded, small data structures)
- **CPU Usage**: Low (simple mathematical operations)
- **File I/O**: Limited to solution output
- **Rendering**: 60 FPS game loop via Ebiten

### Error Handling

The application implements comprehensive error handling:

- **Input Validation**: Real-time equation format checking
- **Mathematical Errors**: Division by zero, inconsistent systems
- **File System Errors**: Solution directory creation and writing
- **UI Errors**: Font loading, window management

### Educational Features

- **Step Visualization**: Each matrix operation is shown individually
- **Operation Explanation**: Text descriptions accompany visual changes
- **Solution Verification**: Results can be verified against original equations
- **Export Functionality**: Solutions saved for reference and study

## Contributing

This project emphasizes educational value and code clarity. When contributing:

1. Maintain the monolithic structure for simplicity
2. Add comprehensive comments for mathematical operations
3. Ensure all changes preserve educational visualization
4. Test with various equation formats and edge cases

## License

[Add your license information here]

---

**Note**: This documentation was enhanced using insights from DeepWiki AI analysis, demonstrating how AI tools can help developers better understand and document their own code architecture.
