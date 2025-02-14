package main

import (
	"embed"
	"fmt"
	"image/color"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed assets/*.ttf
var assets embed.FS

const (
	screenWidth  = 800
	screenHeight = 600
	fontSize     = 20
)

// Matrix represents a mathematical matrix
type Matrix struct {
	rows int
	cols int
	data [][]float64
}

// Game represents the game state
type Game struct {
	equations        [3]string
	activeInput      int
	solving          bool
	solution         string
	steps            []string
	currentStep      int
	stepDelay        int
	font             font.Face
	submitButton     Button
	errorMsg         string
	scrollY          float64
	targetScrollY    float64
	windowHeight     float64
	targetHeight     float64
	scrollBarGrabbed bool
	scrollBarGrabY   int
	matrix           *Matrix
}

// Button represents a clickable button
type Button struct {
	x, y, w, h int
	text       string
}

func (b *Button) Contains(x, y int) bool {
	return x >= b.x && x < b.x+b.w && y >= b.y && y < b.y+b.h
}

// parseEquation parses an equation string into coefficients
func parseEquation(eq string) ([]float64, error) {
	// Remove all spaces and make lowercase
	eq = strings.ToLower(strings.ReplaceAll(eq, " ", ""))

	// Initialize coefficients [x, y, z, constant]
	coeffs := make([]float64, 4)

	// Split at equals sign
	parts := strings.Split(eq, "=")
	if len(parts) != 2 {
		return nil, fmt.Errorf("equation must contain exactly one '=' sign")
	}

	leftSide := parts[0]
	rightSide := parts[1]

	// Parse right side (constant)
	constant, err := strconv.ParseFloat(rightSide, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid constant on right side")
	}
	coeffs[3] = constant

	// Regular expressions for parsing terms
	termRegex := regexp.MustCompile(`[+-]?\d*\.?\d*[xyz]|[+-]?\d+\.?\d*`)
	terms := termRegex.FindAllString(leftSide, -1)

	for _, term := range terms {
		// Default coefficient is 1 or -1 if only sign is present
		coeff := 1.0
		var variable rune

		// Handle different term formats
		if len(term) > 0 {
			if term[0] == '-' {
				coeff = -1.0
				term = term[1:]
			} else if term[0] == '+' {
				term = term[1:]
			}

			// Extract coefficient if present
			if len(term) > 0 {
				if term[0] >= '0' && term[0] <= '9' || term[0] == '.' {
					numPart := term[:len(term)-1]
					if val, err := strconv.ParseFloat(numPart, 64); err == nil {
						coeff *= val
					}
				}
			}

			// Extract variable
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

func (m *Matrix) GaussianElimination() []string {
	steps := []string{}
	lead := 0

	// Helper function to check if a number is effectively zero
	isZero := func(x float64) bool {
		return math.Abs(x) < 1e-10
	}

	// Helper function to round to a specific precision
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
			steps = append(steps, fmt.Sprintf("Swapped row %d and %d", i+1, r+1))
		}

		if !isZero(m.data[r][lead] - 1) {
			scalar := 1.0 / m.data[r][lead]
			scalar = round(scalar, 5)
			m.MultiplyRow(r, scalar)
			steps = append(steps, fmt.Sprintf("Scaled row %d by %.2f", r+1, scalar))
		}

		for i := 0; i < m.rows; i++ {
			if i != r {
				scalar := -m.data[i][lead]
				if !isZero(scalar) {
					scalar = round(scalar, 5)
					m.AddMultipleOfRow(i, r, scalar)
					steps = append(steps, fmt.Sprintf("Added %.2f times row %d to row %d", scalar, r+1, i+1))
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

func (g *Game) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if g.submitButton.Contains(x, y) {
			g.errorMsg = ""
			g.solving = true
			g.solveEquations()
		}

		for i := 0; i < 3; i++ {
			if y >= 80+i*60 && y <= 120+i*60 && x >= 20 && x <= 420 {
				g.activeInput = i
			}
		}

		windowW, _ := ebiten.WindowSize()
		if x >= windowW-20 && x <= windowW {
			g.scrollBarGrabbed = true
			g.scrollBarGrabY = y
		}
	}

	if g.scrollBarGrabbed {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			_, y := ebiten.CursorPosition()
			_, windowH := ebiten.WindowSize()
			deltaY := float64(y - g.scrollBarGrabY)
			contentHeight := g.getContentHeight()

			scrollFactor := deltaY * (float64(contentHeight) - float64(windowH)) / float64(windowH)
			g.targetScrollY = g.scrollY + scrollFactor

			if g.targetScrollY < 0 {
				g.targetScrollY = 0
			}
			maxScroll := float64(contentHeight) - float64(windowH)
			if maxScroll < 0 {
				maxScroll = 0
			}
			if g.targetScrollY > maxScroll {
				g.targetScrollY = maxScroll
			}

			g.scrollBarGrabY = y
		} else {
			g.scrollBarGrabbed = false
		}
	}

	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		g.targetScrollY -= wheelY * 20
		_, windowH := ebiten.WindowSize()
		contentHeight := g.getContentHeight()

		if g.targetScrollY < 0 {
			g.targetScrollY = 0
		}
		maxScroll := float64(contentHeight) - float64(windowH)
		if maxScroll < 0 {
			maxScroll = 0
		}
		if g.targetScrollY > maxScroll {
			g.targetScrollY = maxScroll
		}
	}

	g.scrollY += (g.targetScrollY - g.scrollY) * 0.2

	g.handleKeyboardInput()

	if g.solving && g.currentStep < len(g.steps) {
		g.stepDelay++
		if g.stepDelay > 30 {
			g.currentStep++
			g.stepDelay = 0
		}
	}

	contentHeight := float64(g.getContentHeight())
	if contentHeight != g.targetHeight {
		g.targetHeight = contentHeight
	}
	g.windowHeight += (g.targetHeight - g.windowHeight) * 0.2
	if math.Abs(g.windowHeight-g.targetHeight) > 0.1 {
		ebiten.SetWindowSize(screenWidth, int(g.windowHeight))
	}

	return nil
}

func (g *Game) getContentHeight() int {
	contentHeight := screenHeight

	if g.solving && g.errorMsg == "" {
		numVisibleSteps := g.currentStep + 1
		if numVisibleSteps > len(g.steps) {
			numVisibleSteps = len(g.steps)
		}

		contentHeight = 320 + (numVisibleSteps * 30)

		if g.solution != "" {
			contentHeight += 60
		}

		contentHeight += 50
	}

	return contentHeight
}

func (g *Game) handleKeyboardInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if len(g.equations[g.activeInput]) > 0 {
			g.equations[g.activeInput] = g.equations[g.activeInput][:len(g.equations[g.activeInput])-1]
		}
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.activeInput = (g.activeInput + 1) % 3
		return
	}

	for k := ebiten.Key0; k <= ebiten.Key9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			if len(g.equations[g.activeInput]) < 30 {
				g.equations[g.activeInput] += strconv.Itoa(int(k - ebiten.Key0))
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		g.equations[g.activeInput] += "x"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyY) {
		g.equations[g.activeInput] += "y"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.equations[g.activeInput] += "z"
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		g.equations[g.activeInput] += "-"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		g.equations[g.activeInput] += "="
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.equations[g.activeInput] += " "
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPeriod) {
		g.equations[g.activeInput] += "."
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySlash) {
		g.equations[g.activeInput] += "+"
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	windowW, windowH := screen.Size()
	screen.Fill(color.RGBA{240, 240, 240, 255})
	scrollOffset := math.Floor(g.scrollY)

	text.Draw(screen, "Gaussian Elimination Solver", g.font, 20, 40, color.Black)
	text.Draw(screen, "Enter equations in the form: 2x + y - z = 8", g.font, 20, 70, color.RGBA{100, 100, 100, 255})

	for i := 0; i < 3; i++ {
		y := 100 + i*60
		ebitenutil.DrawRect(screen, 20, float64(y), 400, 40, color.RGBA{255, 255, 255, 255})
		if i == g.activeInput {
			ebitenutil.DrawRect(screen, 20, float64(y), 400, 40, color.RGBA{200, 200, 255, 255})
		}
		text.Draw(screen, g.equations[i], g.font, 30, y+30, color.Black)
	}

	ebitenutil.DrawRect(screen, float64(g.submitButton.x), float64(g.submitButton.y),
		float64(g.submitButton.w), float64(g.submitButton.h),
		color.RGBA{100, 149, 237, 255})
	text.Draw(screen, g.submitButton.text, g.font,
		g.submitButton.x+20, g.submitButton.y+30, color.White)

	if g.solving && g.errorMsg == "" {
		y := 320 - int(scrollOffset)
		for i := 0; i <= g.currentStep && i < len(g.steps); i++ {
			if y+30 >= 0 && y <= windowH {
				text.Draw(screen, g.steps[i], g.font, 20, y, color.Black)
			}
			y += 30
		}
		if g.solution != "" && y <= windowH {
			text.Draw(screen, g.solution, g.font, 20, y+30, color.RGBA{0, 100, 0, 255})
		}
	}

	if g.errorMsg != "" {
		text.Draw(screen, g.errorMsg, g.font, 20, 300, color.RGBA{255, 0, 0, 255})
	}

	contentHeight := g.getContentHeight()
	if contentHeight > windowH {
		ebitenutil.DrawRect(screen, float64(windowW-12), 0, 12, float64(windowH),
			color.RGBA{200, 200, 200, 255})

		scrollbarHeight := float64(windowH) * float64(windowH) / float64(contentHeight)
		scrollbarY := float64(windowH-int(scrollbarHeight)) * g.scrollY / float64(contentHeight-windowH)

		ebitenutil.DrawRect(screen, float64(windowW-10), scrollbarY, 8, scrollbarHeight,
			color.RGBA{100, 100, 100, 255})
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	contentHeight := screenHeight

	if g.solving && g.errorMsg == "" {
		numVisibleSteps := g.currentStep + 1
		if numVisibleSteps > len(g.steps) {
			numVisibleSteps = len(g.steps)
		}

		contentHeight = 320 + (numVisibleSteps * 30)

		if g.solution != "" {
			contentHeight += 60
		}

		contentHeight += 50
	}

	_, windowH := ebiten.WindowSize()
	if contentHeight > windowH {
		ebiten.SetWindowSize(screenWidth, contentHeight)
	}

	return screenWidth, contentHeight
}

func (g *Game) solveEquations() {
	g.matrix = NewMatrix(3, 4)

	for i := 0; i < 3; i++ {
		if g.equations[i] == "" {
			g.errorMsg = fmt.Sprintf("Please enter equation %d", i+1)
			g.solving = false
			return
		}

		coeffs, err := parseEquation(g.equations[i])
		if err != nil {
			g.errorMsg = fmt.Sprintf("Error in equation %d: %s", i+1, err)
			g.solving = false
			return
		}

		g.matrix.data[i] = coeffs
	}

	g.currentStep = 0
	g.stepDelay = 0

	g.steps = g.matrix.GaussianElimination()

	if math.Abs(g.matrix.data[0][0]) < 1e-10 ||
		math.Abs(g.matrix.data[1][1]) < 1e-10 ||
		math.Abs(g.matrix.data[2][2]) < 1e-10 {
		g.errorMsg = "No unique solution exists"
		g.solving = false
		return
	}

	g.steps = append(g.steps, "\nSolution:")
	g.solution = fmt.Sprintf("x = %.2f, y = %.2f, z = %.2f",
		g.matrix.data[0][3], g.matrix.data[1][3], g.matrix.data[2][3])
}

func loadFont() (font.Face, error) {
	fontData, err := assets.ReadFile("assets/DejaVuSans.ttf")
	if err != nil {
		return nil, fmt.Errorf("failed to load font: %v", err)
	}

	tt, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
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
		font: font,
		submitButton: Button{
			x:    20,
			y:    280,
			w:    120,
			h:    40,
			text: "Solve",
		},
	}
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Gaussian Elimination Solver")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
