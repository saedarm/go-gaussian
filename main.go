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
)

type Matrix struct {
	rows int
	cols int
	data [][]float64
}

type Game struct {
	equations        []string
	activeEquation   int
	solving          bool
	solutionComplete bool
	solution         string
	steps            []string
	currentStep      int
	stepDelay        int
	font             font.Face
	matrix           *Matrix
	width, height    int
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

func (g *Game) getContentHeight() int {
	contentHeight := screenHeight

	if g.solving || g.solutionComplete {
		numVisibleSteps := g.currentStep + 1
		if numVisibleSteps > len(g.steps) {
			numVisibleSteps = len(g.steps)
		}

		contentHeight = 320 + (numVisibleSteps * 45)

		if g.solution != "" {
			contentHeight += 80
		}

		contentHeight += 100
	}

	return contentHeight
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	contentHeight := g.getContentHeight()

	if contentHeight > g.height {
		g.height = contentHeight
		ebiten.SetWindowSize(screenWidth, contentHeight)
	}

	if g.solutionComplete && g.height < contentHeight {
		g.height = contentHeight
	}

	return g.width, g.height
}

func (g *Game) Update() error {
	// Only exit if ESC is pressed
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	g.handleInput()

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

	// Never return an error unless ESC is pressed
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
	screen.Fill(color.RGBA{240, 240, 240, 255})

	text.Draw(screen, "Gaussian Elimination Solver", g.font, 20, 40, color.Black)
	text.Draw(screen, "Enter equations in the form: 2x + y - z = 8", g.font, 20, 70, color.RGBA{100, 100, 100, 255})
	text.Draw(screen, "Press SPACE to solve | ESC to exit", g.font, 20, 90, color.RGBA{100, 100, 100, 255})

	for i := 0; i < 3; i++ {
		y := 100 + i*60
		ebitenutil.DrawRect(screen, 20, float64(y), 400, 40, color.RGBA{255, 255, 255, 255})
		if i == g.activeEquation {
			ebitenutil.DrawRect(screen, 20, float64(y), 400, 40, color.RGBA{200, 200, 255, 255})
		}
		text.Draw(screen, g.equations[i], g.font, 30, y+30, color.Black)
	}

	if g.solving || g.solutionComplete {
		y := 320
		for i := 0; i <= g.currentStep && i < len(g.steps); i++ {
			ebitenutil.DrawRect(screen, 20, float64(y-25), float64(g.width-60), 35, color.RGBA{255, 255, 255, 255})
			text.Draw(screen, g.steps[i], g.font, 30, y, color.Black)
			y += 45
		}

		if g.solution != "" {
			ebitenutil.DrawRect(screen, 20, float64(y-25), float64(g.width-60), 35, color.RGBA{230, 255, 230, 255})
			text.Draw(screen, g.solution, g.font, 30, y, color.RGBA{0, 100, 0, 255})
		}
	}
}

func (g *Game) solve() {
	// Don't start a new solution if we're in the middle of one
	if g.solving && !g.solutionComplete {
		return
	}

	g.matrix = NewMatrix(3, 4)

	// Validate all equations first
	for i := 0; i < 3; i++ {
		if g.equations[i] == "" {
			return
		}
		coeffs, err := parseEquation(g.equations[i])
		if err != nil {
			return
		}
		g.matrix.data[i] = coeffs
	}

	g.currentStep = 0
	g.stepDelay = 0
	g.solving = true
	g.solutionComplete = false

	initialMatrix := g.matrix.GetMatrixString()
	g.steps = g.matrix.GaussianElimination()

	if math.Abs(g.matrix.data[0][0]) < 1e-10 ||
		math.Abs(g.matrix.data[1][1]) < 1e-10 ||
		math.Abs(g.matrix.data[2][2]) < 1e-10 {
		g.solving = false
		return
	}

	g.steps = append(g.steps, "\nSolution:")
	g.solution = fmt.Sprintf("x = %.2f, y = %.2f, z = %.2f",
		g.matrix.data[0][3], g.matrix.data[1][3], g.matrix.data[2][3])

	// File saving doesn't affect the window staying open
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
		equations:        make([]string, 3),
		font:             font,
		width:            screenWidth,
		height:           screenHeight,
		solutionComplete: false,
	}
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Gaussian Elimination Solver")
	ebiten.SetWindowResizable(true)

	if err := ebiten.RunGame(NewGame()); err != nil {
		if err == ebiten.Termination {
			return // Normal termination, don't log it
		}
		log.Fatal(err)
	}
}
