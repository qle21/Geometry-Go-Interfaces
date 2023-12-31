// Project Name Image processing in Go Interfaces
// Project Description : a Go program to draw the geometry (either Rectangle, Circle, or Triangle) on the screen and stored in PPM format
// Name Quang Le
// Netid 660670451
// Date 12/01/2023

// Reference
// https://gabrielgambetta.com/computer-graphics-from-scratch/07-filled-triangles.html
// https://www.redblobgames.com/grids/circle-drawing/
package main

import (
	"errors"
	"fmt"
	"os"
)

// Color represents the RGB values of a color
type Color struct {
	R, G, B int
}

// Point represents a point in 2D space
type Point struct {
	Y, X int
}

// Rectangle represents a rectangle shape
type Rectangle struct {
	LL, UR Point
	C      Color
}

// Circle represents a circle shape
type Circle struct {
	CP Point
	R  int
	C  Color
}

// Triangle represents a triangle shape
type Triangle struct {
	Pt0, Pt1, Pt2 Point
	C             Color
}

// Display represents the logical device or physical display
type Display struct {
	MaxX, MaxY int
	Matrix     [][]Color
}

// Screen interface methods
type Screen interface {
	Initialize(maxX, maxY int)
	GetMaxXY() (int, int)
	DrawPixel(x, y int, c Color) error
	GetPixel(x, y int) (Color, error)
	ClearScreen()
	ScreenShot(f string) error
}

// OutOfBoundsError represents an error when a shape is drawn outside the screen
var OutOfBoundsError = errors.New("geometry out of bounds")

// ColorUnknownError represents an error when an unknown color is used
var ColorUnknownError = errors.New("color unknown")

// Initialize initializes the screen with the given dimensions
func (d *Display) Initialize(maxX, maxY int) {
	d.MaxX = maxX
	d.MaxY = maxY
	d.Matrix = make([][]Color, maxY)
	for i := range d.Matrix {
		d.Matrix[i] = make([]Color, maxX)
	}
}

// GetMaxXY returns the maximum X and Y dimensions of the screen
func (d *Display) GetMaxXY() (int, int) {
	return d.MaxX, d.MaxY
}

// DrawPixel draws a pixel with the given color at the specified location
func (d *Display) DrawPixel(x, y int, c Color) error {
	if x < 0 || x >= d.MaxX || y < 0 || y >= d.MaxY {
		return OutOfBoundsError
	}
	d.Matrix[y][x] = c
	return nil
}

// GetPixel returns the color of the pixel at the specified location
func (d *Display) GetPixel(x, y int) (Color, error) {
	if x < 0 || x >= d.MaxX || y < 0 || y >= d.MaxY {
		return Color{}, OutOfBoundsError
	}
	return d.Matrix[y][x], nil
}

// ClearScreen clears the entire screen by setting each pixel to white
func (d *Display) ClearScreen() {
	for y := 0; y < d.MaxY; y++ {
		for x := 0; x < d.MaxX; x++ {
			d.Matrix[y][x] = Color{255, 255, 255} // Set each pixel to white
		}
	}
}

// ScreenShot saves the screen as a .ppm file
func (d *Display) ScreenShot(f string) error {
	file, err := os.Create(f + ".ppm")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "P3\n%d %d\n255\n", d.MaxX, d.MaxY)
	if err != nil {
		return err
	}

	for y := 0; y < d.MaxY; y++ {
		for x := 0; x < d.MaxX; x++ {
			c := d.Matrix[y][x]
			_, err := fmt.Fprintf(file, "%d %d %d ", c.R, c.G, c.B)
			if err != nil {
				return err
			}
		}
		_, err := fmt.Fprintln(file)
		if err != nil {
			return err
		}
	}

	return nil
}

// Interpolate performs linear interpolation between two values code from pdf link
func interpolate(l0, d0, l1, d1 int) []int {
	a := float64(d1-d0) / float64(l1-l0)
	d := float64(d0)

	var values []int
	count := l1 - l0 + 1
	for count > 0 {
		values = append(values, int(d))
		d = d + a
		count--
	}
	return values
}

// Draw draws the filled-in shape on the screen code from pdf link
func (tri Triangle) Draw(scn Screen) error {
	if outOfBounds(tri.Pt0, scn) || outOfBounds(tri.Pt1, scn) || outOfBounds(tri.Pt2, scn) {
		return OutOfBoundsError
	}
	if colorUnknown(tri.C) {
		return ColorUnknownError
	}

	y0 := tri.Pt0.Y
	y1 := tri.Pt1.Y
	y2 := tri.Pt2.Y

	// Sort the points so that y0 <= y1 <= y2 
	if y1 < y0 {
		tri.Pt1, tri.Pt0 = tri.Pt0, tri.Pt1
	}
	if y2 < y0 {
		tri.Pt2, tri.Pt0 = tri.Pt0, tri.Pt2
	}
	if y2 < y1 {
		tri.Pt2, tri.Pt1 = tri.Pt1, tri.Pt2
	}

	x0, y0 := tri.Pt0.X, tri.Pt0.Y
	x1, y1 := tri.Pt1.X, tri.Pt1.Y
	x2, y2 := tri.Pt2.X, tri.Pt2.Y

	x01 := interpolate(y0, x0, y1, x1)
	x12 := interpolate(y1, x1, y2, x2)
	x02 := interpolate(y0, x0, y2, x2)

	// Concatenate the short sides
	x012 := append(x01[:len(x01)-1], x12...)

	// Determine which is left and which is right
	var xLeft, xRight []int
	m := len(x012) / 2
	if x02[m] < x012[m] {
		xLeft = x02
		xRight = x012
	} else {
		xLeft = x012
		xRight = x02
	}

	// Draw the horizontal segments
	for y := y0; y <= y2; y++ {
		for x := xLeft[y-y0]; x <= xRight[y-y0]; x++ {
			scn.DrawPixel(x, y, tri.C)
		}
	}
	return nil
}

// OutOfBounds checks if a point is outside the screen boundaries
func outOfBounds(p Point, scn Screen) bool {
	maxX, maxY := scn.GetMaxXY()
	return p.X < 0 || p.X >= maxX || p.Y < 0 || p.Y >= maxY
}

// ColorUnknown checks if a color is unknown
func colorUnknown(c Color) bool {
	colorMap := map[Color]bool{
		{255, 0, 0}:     true, // red
		{0, 255, 0}:     true, // green
		{0, 0, 255}:     true, // blue
		{255, 255, 0}:   true, // yellow
		{255, 164, 0}:   true, // orange
		{128, 0, 128}:   true, // purple
		{165, 42, 42}:   true, // brown
		{0, 0, 0}:       true, // black
		{255, 255, 255}: true, // white
	}

	_, ok := colorMap[c]
	return !ok
}

// Draw draws the filled-in shape on the screen
func (circ Circle) Draw(scn Screen) error {
	if outOfBounds(circ.CP, scn) {
		return OutOfBoundsError
	}
	if colorUnknown(circ.C) {
		return ColorUnknownError
	}

	x0, y0 := circ.CP.X, circ.CP.Y
	radius := circ.R

	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				scn.DrawPixel(x0+x, y0+y, circ.C)
			}
		}
	}

	return nil
}

// Draw draws the filled-in shape on the screen
func (rect Rectangle) Draw(scn Screen) error {
	if outOfBounds(rect.LL, scn) || outOfBounds(rect.UR, scn) {
		return OutOfBoundsError
	}
	if colorUnknown(rect.C) {
		return ColorUnknownError
	}

	for y := rect.LL.Y; y <= rect.UR.Y; y++ {
		for x := rect.LL.X; x <= rect.UR.X; x++ {
			scn.DrawPixel(x, y, rect.C)
		}
	}

	return nil
}

// declare display struct
var display Display

// declare color
var green = Color{0, 255, 0}
var yellow = Color{255, 255, 0}
var red = Color{255, 0, 0}
var unknow = Color{102, 102, 102}

func main() {

	fmt.Println("starting ...")
	display.Initialize(1024, 1024)
	display.ClearScreen()

	rect := Rectangle{Point{100, 300}, Point{600, 900}, red} // red
	err := rect.Draw(&display)
	if err != nil {
		fmt.Println("rect:", err)
	}

	rect2 := Rectangle{Point{0, 0}, Point{100, 1024}, green} // green
	err = rect2.Draw(&display)
	if err != nil {
		fmt.Println("rect2:", err)
	}

	rect3 := Rectangle{Point{0, 0}, Point{100, 1022}, unknow} // unknown color
	err = rect3.Draw(&display)
	if err != nil {
		fmt.Println("rect3:", err)
	}

	circ := Circle{Point{500, 500}, 200, green} // green
	err = circ.Draw(&display)
	if err != nil {
		fmt.Println("circ:", err)
	}

	tri := Triangle{Point{100, 100}, Point{600, 300}, Point{859, 850}, yellow} // yellow
	err = tri.Draw(&display)
	if err != nil {
		fmt.Println("tri:", err)
	}

	display.ScreenShot("output")
}
