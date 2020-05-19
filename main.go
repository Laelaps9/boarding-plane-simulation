package main

import(
	"image"
	"os"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return pixel.PictureDataFromImage(img), nil
}

func run() {
	y := 300.
	// Specify configuration window
	cfg := pixelgl.WindowConfig {
		Title: "Plane Simulator",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync: true,
	}
	// Create a new window
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Load airplane image
	plane, err := loadPicture("plane_top.png")
	if err != nil {
		panic(err)
	}

	// Create sprite for the airplane image
	sprite := pixel.NewSprite(plane, plane.Bounds())
	
	imd := imdraw.New(nil)

	imd.Color = colornames.Limegreen
	// First entrance x: 163, y: 300
	// Y mid: 420

	for !win.Closed() {
		win.Clear(colornames.Black)
		// Draw sprite in the center of the window
		sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		if y < 420 {
			y += 4
		}
		imd.Clear()
		imd.Push(pixel.V(163, y))
		imd.Circle(10, 0)
		imd.Draw(win)
		win.Update()
	}

}

func main() {
	pixelgl.Run(run)
}