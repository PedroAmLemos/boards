package main

import (
	"fmt"
	"sync"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

type Line struct {
	Start, End raylib.Vector2
}

// var (
// 	currentLine  *Line
// 	selectedLine *Line
// 	selectedEnd  *raylib.Vector2
// )

type BoardState struct {
	currentLine  *Line
	selectedLine *Line
	selectedEnd  *raylib.Vector2
}

type Board struct {
	name       string
	lines      []Line
	updateChan chan Line
	newChan    chan Line
	mu         sync.Mutex
	state      BoardState
}

func NewBoard(name string) *Board {
	return &Board{
		name:       name,
		lines:      []Line{},
		updateChan: make(chan Line),
		newChan:    make(chan Line),
	}
}

func (b *Board) Notifier(thisName string, people map[string]string, connectedClients map[string]string) {
	for {
		select {
		case line := <-b.updateChan:
			fmt.Printf("\nLine updated: x1 = %.2f, y1 = %.2f, x2 = %.2f, y2 = %.2f\n> ", line.Start.X, line.Start.Y, line.End.X, line.End.Y)
		case line := <-b.newChan:
			fmt.Printf("\nLine created: x1 = %.2f, y1 = %.2f, x2 = %.2f, y2 = %.2f\n> ", line.Start.X, line.Start.Y, line.End.X, line.End.Y)
			if b.name == "mainBoard" {
				fmt.Printf("\n[log] New line created at %v, sending it to all clients\n> ", b.name)
				for _, ip := range connectedClients {
					response, err := unicast(thisName, ip, fmt.Sprintf("%v newLine %v %.2f %.2f %.2f %.2f", thisName, thisName, line.Start.X, line.Start.Y, line.End.X, line.End.Y))
					if err != nil {
						fmt.Printf("\n[error] %v\n >", err)
					}
					fmt.Println(string(response))
				}
			} else {
				fmt.Printf("\n[log] New line created at %v, sending it to the owner\n> ", b.name)
				response, err := unicast(thisName, people[b.name], fmt.Sprintf("%v newLine mainBoard %.2f %.2f %.2f %.2f", thisName, line.Start.X, line.Start.Y, line.End.X, line.End.Y))
				if err != nil {
					fmt.Printf("\n[error] %v\n> ", err)
				}
				fmt.Printf("[log] Response for newLine: %v\n >", string(response))
			}
		}

	}

}

func (b *Board) Start(thisName string, people map[string]string, isBoard *bool, connectedClients map[string]string) {
	raylib.InitWindow(1280, 720, b.name)
	raylib.SetTargetFPS(60)

	go b.Notifier(thisName, people, connectedClients)

	fmt.Print("> ")
	for !raylib.WindowShouldClose() {
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
	*isBoard = false
	fmt.Print("> ")
}

func (b *Board) AddLine(newLine Line) {
	b.mu.Lock()
	b.lines = append(b.lines, newLine)
	b.mu.Unlock()
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

// func parseCoords(coords string) (*Line, error) {
// 	parts := strings.Fields(coords)
// 	x1, err := strconv.ParseFloat(parts[0], 64)
// 	if err != nil {
// 		return nil, err
// 	}
// 	y1, err := strconv.ParseFloat(parts[1], 64)
// 	if err != nil {
// 		return nil, err
// 	}
// 	x2, err := strconv.ParseFloat(parts[2], 64)
// 	if err != nil {
// 		return nil, err
// 	}
// 	y2, err := strconv.ParseFloat(parts[3], 64)
// 	if err != nil {
// 		return nil, err
// 	}
// 	newLine := Line{
// 		Start: raylib.Vector2{X: float32(x1), Y: float32(y1)},
// 		End:   raylib.Vector2{X: float32(x2), Y: float32(y2)},
// 	}
// 	return &newLine, nil
// }

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
