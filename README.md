# Gaussian Elimination Solver

An interactive GUI application built with Go and Ebiten that solves systems of linear equations using Gaussian Elimination. The application provides step-by-step visualization of the solution process.

## Features

- Interactive equation input for 3x3 systems
- Real-time parsing and validation of equations
- Step-by-step animated solution process
- Smooth scrolling for long solutions
- Dynamic window resizing
- Error handling and validation
- Clean, modern interface

## Prerequisites

- Go 1.16 or later
- DejaVu Sans font

## Installation

1. Clone the repository:
```bash
git clone [your-repository-url]
cd gaussian-elimination-solver
```

2. Install dependencies:
```bash
go mod init gaussian-solver
go get github.com/hajimehoshi/ebiten/v2
go get golang.org/x/image/font/opentype
```

3. Create an assets directory and add the required font:
```bash
mkdir assets
# Copy DejaVuSans.ttf to the assets directory
```

## Building and Running

Build and run the application:
```bash
go run main.go
```

Or build an executable:
```bash
go build -o gaussian-solver
./gaussian-solver  # or gaussian-solver.exe on Windows
```

## Usage

1. Enter your system of equations in the form:
   - ax + by + cz = d
   - Example: 2x + y - z = 8

2. Input Format:
   - Use x, y, z for variables
   - Use +/- for operators
   - Coefficients can be integers or decimals
   - Each equation must contain one equals sign

3. Controls:
   - Tab/Enter: Move between input fields
   - Mouse: Click input fields or scroll solution
   - Solve button: Start calculation
   - Scrollbar: Navigate long solutions

## Example System

```
2x + y - z = 8
x - y = -3
-x + 2y + 2z = -11
```

## Error Handling

The application handles various error cases:
- Empty equations
- Invalid equation format
- Systems with no unique solution
- Invalid coefficients
- Missing equals signs

## Technical Details

- Built with Ebiten game engine
- Uses Gaussian Elimination algorithm
- Handles floating-point precision issues
- Implements smooth animations and scrolling
- Real-time equation parsing and validation

## Architecture

The application is structured into several key components:
- Matrix operations and Gaussian Elimination logic
- Equation parsing and validation
- Interactive GUI handling
- Real-time animation system
- Window management and event handling

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

[Your chosen license]

## Acknowledgments

- Ebiten game engine
- DejaVu Sans font project
- Gaussian Elimination algorithm
