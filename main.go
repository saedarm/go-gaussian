package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth  = 800
	screenHeight = 600
	fontSize     = 20
	displayTime  = 60 * 30 // 30 seconds at 60 FPS
	minWidth     = 800
	minHeight    = 600
)

type Matrix struct {
	rows int
	cols int
	data [][]float64
}

type Game struct {
	equations           []string
	activeEquation      int
	solving             bool
	solutionComplete    bool
	solution            string
	steps               []string
	currentStep         int
	stepDelay           int
	font                font.Face
	matrix              *Matrix
	width, height       int
	errorMsg            string
	isRunning           bool
	solutionTimer       int
	keepWindowOpen      bool
	ShowExitPrompt      bool
	solutionDisplayDone bool
}

// Matrix operations
func NewMatrix(rows, cols int) *Matrix {
	data := make([][]float64, rows)
	for i := range data {
		data[i] = make([]float64, cols)
	}
	return &Matrix{
		rows: rows,
		cols: cols,
		data: data,
	}
}

func (m *Matrix) SwapRows(row1, row2 int) {
	m.data[row1], m.data[row2] = m.data[row2], m.data[row1]
}

func (m *Matrix) MultiplyRow(row int, scalar float64) {
	for j := 0; j < m.cols; j++ {
		m.data[row][j] *= scalar
	}
}

func (m *Matrix) AddMultipleOfRow(targetRow, sourceRow int, scalar float64) {
	for j := 0; j < m.cols; j++ {
		m.data[targetRow][j] += scalar * m.data[sourceRow][j]
	}
}

func (m *Matrix) GetMatrixString() string {
	var result strings.Builder
	for i := 0; i < m.rows; i++ {
		result.WriteString(fmt.Sprintf("[%.2f %.2f %.2f | %.2f]\n",
			m.data[i][0], m.data[i][1], m.data[i][2], m.data[i][3]))
	}
	return result.String()
}

func (m *Matrix) GaussianElimination() []string {
	steps := []string{}
	lead := 0

	isZero := func(x float64) bool {
		return math.Abs(x) < 1e-10
	}

	round := func(x float64, precision int) float64 {
		multiplier := math.Pow(10, float64(precision))
		return math.Round(x*multiplier) / multiplier
	}

	steps = append(steps, "Starting Gaussian Elimination...")

	for r := 0; r < m.rows; r++ {
		if lead >= m.cols {
			return steps
		}

		i := r
		for isZero(m.data[i][lead]) {
			i++
			if i == m.rows {
				i = r
				lead++
				if lead == m.cols {
					return steps
				}
			}
		}

		if i != r {
			m.SwapRows(i, r)
			steps = append(steps, fmt.Sprintf("L%d ↔ L%d", i+1, r+1))
		}

		if !isZero(m.data[r][lead] - 1) {
			scalar := 1.0 / m.data[r][lead]
			scalar = round(scalar, 5)
			m.MultiplyRow(r, scalar)
			steps = append(steps, fmt.Sprintf("L%d → %.2fL%d", r+1, scalar, r+1))
		}

		for i := 0; i < m.rows; i++ {
			if i != r {
				scalar := -m.data[i][lead]
				if !isZero(scalar) {
					scalar = round(scalar, 5)
					m.AddMultipleOfRow(i, r, scalar)
					if scalar == -1 {
						steps = append(steps, fmt.Sprintf("L%d + L%d → L%d", i+1, r+1, i+1))
					} else {
						steps = append(steps, fmt.Sprintf("L%d + %.2fL%d → L%d", i+1, scalar, r+1, i+1))
					}
				}
			}
		}

		for i := 0; i < m.rows; i++ {
			for j := 0; j < m.cols; j++ {
				m.data[i][j] = round(m.data[i][j], 5)
			}
		}

		lead++
	}

	return steps
}

// Equation parsing
func parseEquation(eq string) ([]float64, error) {
	eq = strings.ToLower(strings.ReplaceAll(eq, " ", ""))
	coeffs := make([]float64, 4)

	parts := strings.Split(eq, "=")
	if len(parts) != 2 {
		return nil, fmt.Errorf("equation must contain exactly one '=' sign")
	}

	leftSide := parts[0]
	rightSide := parts[1]

	constant, err := strconv.ParseFloat(rightSide, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid constant on right side")
	}
	coeffs[3] = constant

	termRegex := regexp.MustCompile(`[+-]?\d*\.?\d*[xyz]|[+-]?\d+\.?\d*`)
	terms := termRegex.FindAllString(leftSide, -1)

	for _, term := range terms {
		coeff := 1.0
		var variable rune

		if len(term) > 0 {
			if term[0] == '-' {
				coeff = -1.0
				term = term[1:]
			} else if term[0] == '+' {
				term = term[1:]
			}

			if len(term) > 0 {
				if term[0] >= '0' && term[0] <= '9' || term[0] == '.' {
					numPart := term[:len(term)-1]
					if val, err := strconv.ParseFloat(numPart, 64); err == nil {
						coeff *= val
					}
				}
			}

			if len(term) > 0 {
				variable = rune(term[len(term)-1])
				switch variable {
				case 'x':
					coeffs[0] += coeff
				case 'y':
					coeffs[1] += coeff
				case 'z':
					coeffs[2] += coeff
				}
			}
		}
	}

	return coeffs, nil
}

// Game methods
func (g *Game) getContentHeight() int {
	contentHeight := minHeight // Start with minimum height

	if g.solving || g.solutionComplete {
		numVisibleSteps := g.currentStep + 1
		if numVisibleSteps > len(g.steps) {
			numVisibleSteps = len(g.steps)
		}

		height := 320 + (numVisibleSteps * 45)

		if g.solution != "" {
			height += 80
		}

		height += 100

		if height > contentHeight {
			contentHeight = height
		}
	}

	// Ensure we never return less than minHeight
	if contentHeight < minHeight {
		contentHeight = minHeight
	}

	return contentHeight
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	contentHeight := g.getContentHeight()

	// Ensure minimum dimensions
	width := minWidth
	height := contentHeight

	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	// Update game dimensions
	g.width = width
	g.height = height

	return width, height
}
func (g *Game) Update() error {
	if !g.isRunning {
		return nil
	}

	// Handle window closing event
	if ebiten.IsWindowBeingClosed() {
		if g.solutionDisplayDone {
			g.ShowExitPrompt = true
			g.isRunning = false
			return ebiten.Termination
		}
		ebiten.SetWindowClosingHandled(true)
		return nil
	}

	// Only exit if ESC is pressed and solution is complete
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && g.solutionDisplayDone {
		g.isRunning = false
		return ebiten.Termination
	}

	g.handleInput()

	// Handle solution timer
	if g.keepWindowOpen {
		if g.solutionTimer > 0 {
			g.solutionTimer--
			if g.solutionTimer <= 0 {
				g.solutionDisplayDone = true
				g.keepWindowOpen = false
			}
		}
	}

	// Continue animation even after solution is complete
	if g.solving && g.currentStep < len(g.steps) {
		g.stepDelay++
		if g.stepDelay > 30 {
			g.currentStep++
			g.stepDelay = 0
			if g.currentStep >= len(g.steps) {
				g.solutionComplete = true
			}
		}
	}

	return nil
}

func (g *Game) handleInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if len(g.equations[g.activeEquation]) > 0 {
			g.equations[g.activeEquation] = g.equations[g.activeEquation][:len(g.equations[g.activeEquation])-1]
		}
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.activeEquation = (g.activeEquation + 1) % 3
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.solve()
		return
	}

	for k := ebiten.Key0; k <= ebiten.Key9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			g.equations[g.activeEquation] += strconv.Itoa(int(k - ebiten.Key0))
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		g.equations[g.activeEquation] += "x"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyY) {
		g.equations[g.activeEquation] += "y"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.equations[g.activeEquation] += "z"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		g.equations[g.activeEquation] += "-"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		g.equations[g.activeEquation] += "="
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySlash) {
		g.equations[g.activeEquation] += "+"
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Get actual screen dimensions
	actualWidth, actualHeight := screen.Size()

	// Fill background
	screen.Fill(color.RGBA{240, 240, 240, 255})

	// Draw title and instructions
	text.Draw(screen, "Gaussian Elimination Solver", g.font, 20, 40, color.Black)
	text.Draw(screen, "Enter equations in the form: 2x + y - z = 8", g.font, 20, 70, color.RGBA{100, 100, 100, 255})
	text.Draw(screen, "Press SPACE to solve | ESC to exit", g.font, 20, 90, color.RGBA{100, 100, 100, 255})

	// Draw equation input fields
	for i := 0; i < 3; i++ {
		y := 100 + i*60
		ebitenutil.DrawRect(screen, 20, float64(y), 400, 40, color.RGBA{255, 255, 255, 255})
		if i == g.activeEquation {
			ebitenutil.DrawRect(screen, 20, float64(y), 400, 40, color.RGBA{200, 200, 255, 255})
		}
		text.Draw(screen, g.equations[i], g.font, 30, y+30, color.Black)
	}

	// Draw solution steps
	if g.solving || g.solutionComplete {
		y := 320
		for i := 0; i <= g.currentStep && i < len(g.steps); i++ {
			ebitenutil.DrawRect(screen, 20, float64(y-25), float64(actualWidth-60), 35, color.RGBA{255, 255, 255, 255})
			text.Draw(screen, g.steps[i], g.font, 30, y, color.Black)
			y += 45
		}

		if g.solution != "" {
			ebitenutil.DrawRect(screen, 20, float64(y-25), float64(actualWidth-60), 35, color.RGBA{230, 255, 230, 255})
			text.Draw(screen, g.solution, g.font, 30, y, color.RGBA{0, 100, 0, 255})
		}
	}

	// Draw error message if any
	if g.errorMsg != "" {
		text.Draw(screen, g.errorMsg, g.font, 20, 300, color.RGBA{255, 0, 0, 255})
	}

	// Draw exit prompt if showing
	if g.ShowExitPrompt {
		// Create overlay with safe dimensions
		overlayWidth := actualWidth
		if overlayWidth < 1 {
			overlayWidth = 1
		}
		overlayHeight := actualHeight
		if overlayHeight < 1 {
			overlayHeight = 1
		}

		overlay := ebiten.NewImage(overlayWidth, overlayHeight)
		overlay.Fill(color.RGBA{0, 0, 0, 128})
		opt := &ebiten.DrawImageOptions{}
		screen.DrawImage(overlay, opt)

		msg := "Solution complete. Press ESC to exit."
		bound := text.BoundString(g.font, msg)
		x := (overlayWidth - bound.Dx()) / 2
		y := (overlayHeight - bound.Dy()) / 2
		text.Draw(screen, msg, g.font, x, y, color.White)
	}
}

func (g *Game) solve() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from error in solve: %v", r)
			g.errorMsg = "An error occurred while solving"
			g.solving = true
			g.solutionComplete = true
		}
	}()

	g.matrix = NewMatrix(3, 4)
	g.errorMsg = ""

	// Validate equations
	for i := 0; i < 3; i++ {
		if g.equations[i] == "" {
			g.errorMsg = fmt.Sprintf("Please enter equation %d", i+1)
			return
		}
		coeffs, err := parseEquation(g.equations[i])
		if err != nil {
			g.errorMsg = fmt.Sprintf("Error in equation %d: %s", i+1, err)
			return
		}
		g.matrix.data[i] = coeffs
	}

	g.currentStep = 0
	g.stepDelay = 0
	g.solving = true
	g.solutionComplete = false
	g.solutionDisplayDone = false
	g.ShowExitPrompt = false

	initialMatrix := g.matrix.GetMatrixString()
	g.steps = g.matrix.GaussianElimination()

	if math.Abs(g.matrix.data[0][0]) < 1e-10 ||
		math.Abs(g.matrix.data[1][1]) < 1e-10 ||
		math.Abs(g.matrix.data[2][2]) < 1e-10 {
		g.errorMsg = "No unique solution exists"
		return
	}

	g.steps = append(g.steps, "\nSolution:")
	g.solution = fmt.Sprintf("x = %.2f, y = %.2f, z = %.2f",
		g.matrix.data[0][3], g.matrix.data[1][3], g.matrix.data[2][3])

	// Start the solution timer
	g.solutionTimer = displayTime
	g.keepWindowOpen = true

	// Handle file operations
	err := os.MkdirAll("solutions", 0755)
	if err != nil {
		log.Printf("Error creating solutions directory: %v", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := filepath.Join("solutions", fmt.Sprintf("gaussian_solution_%s.txt", timestamp))

	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		return
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("Solution generated at: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	f.WriteString("Input Equations:\n")
	for i, eq := range g.equations {
		f.WriteString(fmt.Sprintf("Equation %d: %s\n", i+1, eq))
	}

	f.WriteString("\nInitial Matrix:\n")
	f.WriteString(initialMatrix)

	f.WriteString("\nSolution Steps:\n")
	for _, step := range g.steps {
		f.WriteString(step + "\n")
	}

	f.WriteString("\nFinal Matrix:\n")
	f.WriteString(g.matrix.GetMatrixString())

	f.WriteString("\n" + g.solution + "\n")
}

func loadFont() (font.Face, error) {
	tt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	return opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

func NewGame() *Game {
	font, err := loadFont()
	if err != nil {
		log.Fatal(err)
	}

	return &Game{
		equations:           make([]string, 3),
		font:                font,
		width:               minWidth,
		height:              minHeight,
		solutionComplete:    false,
		solving:             false,
		isRunning:           true,
		solutionTimer:       0,
		keepWindowOpen:      false,
		ShowExitPrompt:      false,
		solutionDisplayDone: false,
	}
}

func main() {
	ebiten.SetWindowSize(minWidth, minHeight)
	ebiten.SetWindowTitle("Gaussian Elimination Solver")
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowClosingHandled(true)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		if err == ebiten.Termination {
			os.Exit(0) // Clean exit
		}
		log.Printf("Game error: %v", err)
	}
}
