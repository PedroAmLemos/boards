package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

type Line struct {
	Start, End raylib.Vector2
}

func (l Line) String() string {
	return fmt.Sprintf("%f %f %f %f", l.Start.X, l.Start.Y, l.End.X, l.End.Y)
}

type BoardState struct {
	currentLine  *Line
	selectedLine *Line
	selectedEnd  *raylib.Vector2
}

type Clients map[string]struct{}

type Board struct {
	name             string
	lines            []Line
	updateChan       chan Line
	newChan          chan Line
	mu               sync.Mutex
	state            BoardState
	connectedClients Clients
}

func NewBoard(name string) *Board {
	return &Board{
		name:             name,
		lines:            []Line{},
		updateChan:       make(chan Line),
		newChan:          make(chan Line),
		connectedClients: make(Clients),
	}
}

func (b *Board) Notifier(nodes map[string]*Node) {
	for {
		select {
		case line := <-b.updateChan:
			fmt.Printf(
				"\nLine updated: x1 = %.2f, y1 = %.2f, x2 = %.2f, y2 = %.2f\n> ",
				line.Start.X,
				line.Start.Y,
				line.End.X,
				line.End.Y,
			)
			if b.name == "mainBoard" {
				for client := range b.connectedClients {
					unicast(
						nodes,
						client,
						fmt.Sprintf(
							"updateline %v %f %f %f %f",
							nodes["thisNode"].name,
							line.Start.X,
							line.Start.Y,
							line.End.X,
							line.End.Y,
						),
					)
				}
			} else {
				unicast(nodes, b.name, fmt.Sprintf("updateline mainBoard %f %f %f %f", line.Start.X, line.Start.Y, line.End.X, line.End.Y))
			}
		case line := <-b.newChan:
			fmt.Printf(
				"\nLine created: x1 = %.2f, y1 = %.2f, x2 = %.2f, y2 = %.2f\n> ",
				line.Start.X,
				line.Start.Y,
				line.End.X,
				line.End.Y,
			)
			if b.name == "mainBoard" {
				for client := range b.connectedClients {
					unicast(
						nodes,
						client,
						fmt.Sprintf(
							"newline %v %f %f %f %f",
							nodes["thisNode"].name,
							line.Start.X,
							line.Start.Y,
							line.End.X,
							line.End.Y,
						),
					)
				}
			} else {
				unicast(nodes, b.name, fmt.Sprintf("newline mainBoard %f %f %f %f", line.Start.X, line.Start.Y, line.End.X, line.End.Y))
			}
		}
		fmt.Printf("\n> ")
	}
}

func (b *Board) Start(nodes map[string]*Node, activeBoard *bool) {
	raylib.InitWindow(1280, 720, b.name)
	raylib.SetTargetFPS(60)

	go b.Notifier(nodes)

	fmt.Print("> ")
	for !raylib.WindowShouldClose() && *activeBoard {
		b.mu.Lock()
		b.HandleInput()
		b.mu.Unlock()

		b.mu.Lock()
		raylib.BeginDrawing()
		raylib.ClearBackground(raylib.RayWhite)
		b.DrawLines()
		raylib.EndDrawing()
		b.mu.Unlock()
	}

	raylib.CloseWindow()
	fmt.Print("> ")
}

func (b *Board) AddLine(newLine Line) {
	b.lines = append(b.lines, newLine)
}

func (b *Board) UpdateLine(updatedLine Line) {
	for i := range b.lines {
		if b.lines[i].Start == updatedLine.Start || b.lines[i].End == updatedLine.End {
			b.lines[i] = updatedLine
			break
		}
	}
}

func (b *Board) HandleInput() {
	mousePos := raylib.GetMousePosition()
	if raylib.IsMouseButtonPressed(raylib.MouseLeftButton) {
		if b.state.selectedLine == nil {
			for i := range b.lines {
				if raylib.CheckCollisionPointCircle(mousePos, b.lines[i].Start, 5) {
					b.state.selectedLine = &b.lines[i]
					b.state.selectedEnd = &b.state.selectedLine.Start
					break
				} else if raylib.CheckCollisionPointCircle(mousePos, b.lines[i].End, 5) {
					b.state.selectedLine = &b.lines[i]
					b.state.selectedEnd = &b.state.selectedLine.End
					break
				}
			}
			if b.state.selectedLine == nil {
				b.state.currentLine = &Line{Start: mousePos, End: mousePos}
			}
		}
	} else if raylib.IsMouseButtonDown(raylib.MouseLeftButton) {
		if b.state.currentLine != nil {
			b.state.currentLine.End = mousePos
		} else if b.state.selectedLine != nil {
			*b.state.selectedEnd = mousePos
		}
	} else if raylib.IsMouseButtonReleased(raylib.MouseLeftButton) {
		if b.state.currentLine != nil {
			b.lines = append(b.lines, *b.state.currentLine)
			b.newChan <- *b.state.currentLine
			b.state.currentLine = nil
		} else if b.state.selectedLine != nil {
			b.updateChan <- *b.state.selectedLine
			b.state.selectedLine = nil
			b.state.selectedEnd = nil
		}
	}
}

func (b *Board) DrawLines() {
	for _, line := range b.lines {
		raylib.DrawLineEx(line.Start, line.End, 2, raylib.DarkGray)
		raylib.DrawCircleV(line.Start, 5, raylib.Red)
		raylib.DrawCircleV(line.End, 5, raylib.Red)
	}
	if b.state.currentLine != nil {
		raylib.DrawLineEx(b.state.currentLine.Start, b.state.currentLine.End, 2, raylib.DarkGray)
		raylib.DrawCircleV(b.state.currentLine.Start, 5, raylib.Red)
		raylib.DrawCircleV(b.state.currentLine.End, 5, raylib.Red)
	}
}

func (b *Board) GetLines() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	var linesString string
	linesString += fmt.Sprintf("%d\n", len(b.lines))
	for _, line := range b.lines {
		linesString += fmt.Sprintf("%.2f %.2f %.2f %.2f\n", line.Start.X, line.Start.Y, line.End.X, line.End.Y)
	}
	return linesString
}

func (b *Board) CloseBoard() {
	raylib.CloseWindow()
}

func parseLine(line []string) Line {
	x1, err := strconv.ParseFloat(line[0], 64)
	if err != nil {
		fmt.Println("[error] Error parsing x1:", err)
	}
	y1, err := strconv.ParseFloat(line[1], 64)
	if err != nil {
		fmt.Println("[error] Error parsing y1:", err)
	}
	x2, err := strconv.ParseFloat(line[2], 64)
	if err != nil {
		fmt.Println("[error] Error parsing x2:", err)
	}
	y2, err := strconv.ParseFloat(line[3], 64)
	if err != nil {
		fmt.Println("[error] Error parsing y2:", err)
	}
	return Line{
		Start: raylib.Vector2{X: float32(x1), Y: float32(y1)},
		End:   raylib.Vector2{X: float32(x2), Y: float32(y2)},
	}
}

func parseLines(lines string) []Line {
	var result []Line

	linesArr := strings.Split(strings.TrimSpace(lines), "\n")

	numLines, err := strconv.Atoi(linesArr[0])
	if err != nil {
		fmt.Println("[error] Error parsing number of lines:", err)
		return result
	}

	for i := 1; i <= numLines; i++ {
		lineArr := strings.Split(linesArr[i], " ")
		if len(lineArr) != 4 {
			fmt.Println("[error] Invalid line format:", linesArr[i])
			continue
		}
		x1, err := strconv.ParseFloat(lineArr[0], 64)
		if err != nil {
			fmt.Println("[error] Error parsing x1:", err)
			continue
		}
		y1, err := strconv.ParseFloat(lineArr[1], 64)
		if err != nil {
			fmt.Println("[error] Error parsing y1:", err)
			continue
		}
		x2, err := strconv.ParseFloat(lineArr[2], 64)
		if err != nil {
			fmt.Println("[error] Error parsing x2:", err)
			continue
		}
		y2, err := strconv.ParseFloat(lineArr[3], 64)
		if err != nil {
			fmt.Println("[error] Error parsing y2:", err)
			continue
		}
		result = append(result, Line{
			Start: raylib.Vector2{X: float32(x1), Y: float32(y1)},
			End:   raylib.Vector2{X: float32(x2), Y: float32(y2)},
		})
	}

	return result
}
