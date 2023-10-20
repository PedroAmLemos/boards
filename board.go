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

var (
	lines          []Line
	currentLine    *Line
	selectedLine   *Line
	selectedEnd    *raylib.Vector2
	mu             sync.Mutex
	updateNotifier chan Line
	newNotifier    chan Line
)

func parseCoords(coords string) (*Line, error) {
	parts := strings.Fields(coords)
	x1, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, err
	}
	y1, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, err
	}
	x2, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil, err
	}
	y2, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return nil, err
	}
	newLine := Line{
		Start: raylib.Vector2{X: float32(x1), Y: float32(y1)},
		End:   raylib.Vector2{X: float32(x2), Y: float32(y2)},
	}
	return &newLine, nil
}

func createBoard() {
	raylib.InitWindow(1280, 720, "Draw Lines")
	raylib.SetTargetFPS(60)
	updateNotifier = make(chan Line)
	newNotifier = make(chan Line)

	go lineNotifier()

	fmt.Print("> ")
	for !raylib.WindowShouldClose() {
		mu.Lock()
		handleInput()
		mu.Unlock()

		mu.Lock()
		raylib.BeginDrawing()
		raylib.ClearBackground(raylib.RayWhite)
		drawLines()
		raylib.EndDrawing()
		mu.Unlock()
	}

	raylib.CloseWindow()
	board = false
	fmt.Print("> ")
}

func createLine(line Line) {
	mu.Lock()
	defer mu.Unlock()
	lines = append(lines, line)
}

func handleInput() {
	mousePos := raylib.GetMousePosition()
	if raylib.IsMouseButtonPressed(raylib.MouseLeftButton) {
		if selectedLine == nil {
			for i := range lines {
				if raylib.CheckCollisionPointCircle(mousePos, lines[i].Start, 5) {
					selectedLine = &lines[i]
					selectedEnd = &selectedLine.Start
					break
				} else if raylib.CheckCollisionPointCircle(mousePos, lines[i].End, 5) {
					selectedLine = &lines[i]
					selectedEnd = &selectedLine.End
					break
				}
			}
			if selectedLine == nil {
				currentLine = &Line{Start: mousePos, End: mousePos}
			}
		}
	} else if raylib.IsMouseButtonDown(raylib.MouseLeftButton) {
		if currentLine != nil {
			currentLine.End = mousePos
		} else if selectedLine != nil {
			*selectedEnd = mousePos
		}
	} else if raylib.IsMouseButtonReleased(raylib.MouseLeftButton) {
		if currentLine != nil {
			lines = append(lines, *currentLine)
			newNotifier <- *currentLine
			currentLine = nil
		} else if selectedLine != nil {
			updateNotifier <- *selectedLine
			selectedLine = nil
			selectedEnd = nil
		}
	}
}

func drawLines() {
	for _, line := range lines {
		raylib.DrawLineEx(line.Start, line.End, 2, raylib.DarkGray)
		raylib.DrawCircleV(line.Start, 5, raylib.Red)
		raylib.DrawCircleV(line.End, 5, raylib.Red)
	}
	if currentLine != nil {
		raylib.DrawLineEx(currentLine.Start, currentLine.End, 2, raylib.DarkGray)
		raylib.DrawCircleV(currentLine.Start, 5, raylib.Red)
		raylib.DrawCircleV(currentLine.End, 5, raylib.Red)
	}
}

func lineNotifier() {
	for {
		select {
		case line := <-updateNotifier:
			fmt.Printf("\nLine updated: x1 = %.2f, y1 = %.2f, x2 = %.2f, y2 = %.2f\n> ", line.Start.X, line.Start.Y, line.End.X, line.End.Y)
		case line := <-newNotifier:
			fmt.Printf("\nLine created: x1 = %.2f, y1 = %.2f, x2 = %.2f, y2 = %.2f\n> ", line.Start.X, line.Start.Y, line.End.X, line.End.Y)
		}

	}
}

// function to return all the lines as a string, where the first line is the number of lines and each line is a line
func getLines() string {
	mu.Lock()
	defer mu.Unlock()
	var linesString string
	linesString += fmt.Sprintf("%d\n", len(lines))
	for _, line := range lines {
		linesString += fmt.Sprintf("%.2f %.2f %.2f %.2f\n", line.Start.X, line.Start.Y, line.End.X, line.End.Y)
	}
	return linesString
}
