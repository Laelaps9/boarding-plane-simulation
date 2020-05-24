package main

import (
	"fmt"
	"math/rand"
	"time"

	_ "image/png"

	"github.com/bradfitz/slice"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

// State masking
const STANDING = 0
const HANDLING_BAHS = 1
const SITTING = 2

// Order masking
const RANDOM = 0
const BACK_TO_FRONT = 1
const FRONT_TO_BACK = 2
const WINDOW_TO_AISLE = 3

// Global varaibles
var elapsed = 0
var seated = 0
var passengers [144]Passenger

type Passenger struct {
	PosX     int
	PosY     int
	SeatN    int
	SeatL    string
	State    int
	Delay    int
	BagsDone bool
}

func generatePasses(orderFlag int) {

	var seats [144]int
	rows := [6]string{"A", "B", "C", "D", "E", "F"}

	// Create boarding passes
	for i := range seats {
		seats[i] = i
	}

	var j = 0

	// Assign boarding passes
	for i := range passengers {

		if i%24 == 0 && i != 0 {
			j++
		}

		passengers[i].PosX = 0
		passengers[i].PosY = 0

		passengers[i].SeatN = (i % 24) + 1
		passengers[i].SeatL = rows[j]

		passengers[i].State = STANDING
		passengers[i].Delay = 0
		passengers[i].BagsDone = false
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(passengers), func(i, j int) { passengers[i], passengers[j] = passengers[j], passengers[i] })

	if orderFlag == BACK_TO_FRONT {
		slice.Sort(passengers[:], func(i, j int) bool {
			return passengers[i].SeatN > passengers[j].SeatN
		})
	} else if orderFlag == FRONT_TO_BACK {
		slice.Sort(passengers[:], func(i, j int) bool {
			return passengers[i].SeatN < passengers[j].SeatN
		})
	} else if orderFlag == WINDOW_TO_AISLE {
		tmpPassengers := make([]Passenger, 144)
		var i = 0

		for j := range passengers {
			if passengers[j].SeatL == "A" || passengers[j].SeatL == "F" {
				tmpPassengers[i] = passengers[j]
				passengers[j].SeatL = "X"
				i++
			}
		}

		for j := range passengers {
			if passengers[j].SeatL == "B" || passengers[j].SeatL == "E" {
				tmpPassengers[i] = passengers[j]
				passengers[j].SeatL = "X"
				i++
			}
		}

		for j := range passengers {
			if passengers[j].SeatL == "C" || passengers[j].SeatL == "D" {
				tmpPassengers[i] = passengers[j]
				passengers[j].SeatL = "X"
				i++
			}
		}

		for i := range passengers {
			passengers[i] = tmpPassengers[i]
		}

	}

	for i := range passengers {
		passengers[i].PosX = i * -1
	}

}

// Check if a specific coordinate is clear or occupied
func isFree(PosX int, PosY int) bool {
	for i := range passengers {
		if passengers[i].PosX == PosX && passengers[i].PosY == PosY {
			return false
		}
	}
	return true
}

func getPassengerInPosition(PosX int, PosY int) Passenger {
	for i := range passengers {
		if passengers[i].PosX == PosX && passengers[i].PosY == PosY {
			return passengers[i]
		}
	}
	var empty Passenger
	return empty
}

func swapPassengers(walker Passenger, obstructer Passenger) {
	walkerX := walker.PosX
	walkerY := walker.PosY

	walker.PosX = obstructer.PosX
	walker.PosY = obstructer.PosY

	obstructer.PosX = walkerX
	obstructer.PosY = walkerY

	if obstructer.State == SITTING {
		obstructer.State = STANDING
		seated--
	}
}

func board(size int) {

	for {

		fmt.Println(passengers[0:size])

		// Finish once all passengers are seated
		if seated == size {
			break
		}

		for i := range passengers {

			// Walk towards row number on the aisle
			if passengers[i].PosX < passengers[i].SeatN {
				if isFree(passengers[i].PosX+1, passengers[i].PosY) {
					passengers[i].PosX++
				}
				passengers[i].State = STANDING

			} else if passengers[i].PosX > passengers[i].SeatN {
				if isFree(passengers[i].PosX+1, passengers[i].PosY) {
					passengers[i].PosX--
				}
				passengers[i].State = STANDING

			} else {

				// Handle bags once row is approached
				if passengers[i].BagsDone == false && passengers[i].State == STANDING {
					passengers[i].State = HANDLING_BAHS
					passengers[i].Delay = 3
				} else if passengers[i].Delay != 0 && passengers[i].State == HANDLING_BAHS {
					passengers[i].Delay--
				} else if passengers[i].Delay == 0 && passengers[i].State == HANDLING_BAHS {
					passengers[i].BagsDone = true
				}

				// Once bags are handle, attempt to sit down
				if passengers[i].BagsDone == true {

					if passengers[i].SeatL == "A" && passengers[i].PosY > -3 {
						if isFree(passengers[i].PosX, passengers[i].PosY-1) {
							passengers[i].PosY--
						} else {
							// Switch positions with obstructing poassenger
							obstructer := getPassengerInPosition(passengers[i].PosX, passengers[i].PosY-1)
							swapPassengers(passengers[i], obstructer)
						}

						// Sit down when seat if found
					} else if passengers[i].SeatL == "A" && passengers[i].PosY == -3 && passengers[i].State != SITTING {
						passengers[i].State = SITTING
						seated++
					}

					if passengers[i].SeatL == "B" && passengers[i].PosY > -2 {
						if isFree(passengers[i].PosX, passengers[i].PosY-1) {
							passengers[i].PosY--
						} else {
							// Switch positions with obstructing poassenger
							obstructer := getPassengerInPosition(passengers[i].PosX, passengers[i].PosY-1)
							swapPassengers(passengers[i], obstructer)
						}

						// Sit down when seat if found
					} else if passengers[i].SeatL == "B" && passengers[i].PosY == -2 && passengers[i].State != SITTING {
						passengers[i].State = SITTING
						seated++
					}

					if passengers[i].SeatL == "C" && passengers[i].PosY > -1 {
						if isFree(passengers[i].PosX, passengers[i].PosY-1) {
							passengers[i].PosY--
						} else {
							// Switch positions with obstructing poassenger
							obstructer := getPassengerInPosition(passengers[i].PosX, passengers[i].PosY-1)
							swapPassengers(passengers[i], obstructer)
						}

						// Sit down when seat if found
					} else if passengers[i].SeatL == "C" && passengers[i].PosY == -1 && passengers[i].State != SITTING {
						passengers[i].State = SITTING
						seated++
					}

					if passengers[i].SeatL == "D" && passengers[i].PosY < 1 {
						if isFree(passengers[i].PosX, passengers[i].PosY+1) {
							passengers[i].PosY++
						} else {
							// Switch positions with obstructing poassenger
							obstructer := getPassengerInPosition(passengers[i].PosX, passengers[i].PosY+1)
							swapPassengers(passengers[i], obstructer)
						}

						// Sit down when seat if found
					} else if passengers[i].SeatL == "D" && passengers[i].PosY == 1 && passengers[i].State != SITTING {
						passengers[i].State = SITTING
						seated++
					}

					if passengers[i].SeatL == "E" && passengers[i].PosY < 2 {
						if isFree(passengers[i].PosX, passengers[i].PosY+1) {
							passengers[i].PosY++
						} else {
							// Switch positions with obstructing poassenger
							obstructer := getPassengerInPosition(passengers[i].PosX, passengers[i].PosY+1)
							swapPassengers(passengers[i], obstructer)
						}

						// Sit down when seat if found
					} else if passengers[i].SeatL == "E" && passengers[i].PosY == 2 && passengers[i].State != SITTING {
						passengers[i].State = SITTING
						seated++
					}

					if passengers[i].SeatL == "F" && passengers[i].PosY < 3 {
						if isFree(passengers[i].PosX, passengers[i].PosY+1) {
							passengers[i].PosY++
						} else {
							// Switch positions with obstructing poassenger
							obstructer := getPassengerInPosition(passengers[i].PosX, passengers[i].PosY+1)
							swapPassengers(passengers[i], obstructer)
						}

						// Sit down when seat if found
					} else if passengers[i].SeatL == "F" && passengers[i].PosY == 3 && passengers[i].State != SITTING {
						passengers[i].State = SITTING
						seated++
					}
				}

			}

		}

		elapsed++
	}

	fmt.Println("Elapsed:", elapsed, "s")

}

func createWindow() *pixelgl.Window {
	// Specify configuration window
	cfg := pixelgl.WindowConfig{
		Title:  "Plane Boarding Simulator",
		Bounds: pixel.R(0, 0, 1430, 900),
		VSync:  true,
	}
	// Create a new window
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	win.Clear(colornames.Black)

	return win
}

func drawPlane() *imdraw.IMDraw {
	plane := imdraw.New(nil)

	plane.Color = colornames.Lightgray
	plane.Push(pixel.V(0, 900))
	plane.Push(pixel.V(1480, 480))
	plane.Rectangle(0)

	plane.Color = colornames.Red

	// Front Entrance
	plane.Push(pixel.V(10, 482))
	plane.Push(pixel.V(50, 478))
	plane.Rectangle(0)

	// Back Entrance
	plane.Push(pixel.V(1370, 482))
	plane.Push(pixel.V(1410, 478))
	plane.Rectangle(0)

	// Draw seats
	plane.Color = colornames.Darkgray
	for i := 0; i < 3; i++ {
		for j := 1; j <= 24; j++ {
			plane.Push(pixel.V(float64(52*j+40), float64(860-50*i)))
			plane.Push(pixel.V(float64(52*j+80), float64(820-50*i)))
			plane.Rectangle(0)
		}
	}

	for i := 0; i < 3; i++ {
		for j := 1; j <= 24; j++ {
			plane.Push(pixel.V(float64(52*j+40), float64(660-50*i)))
			plane.Push(pixel.V(float64(52*j+80), float64(620-50*i)))
			plane.Rectangle(0)
		}
	}

	return plane
}

func drawLabels() (*text.Text, *text.Text, *text.Text) {
	// Prepare font to print text on screen
	txt := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	// Other labels
	others := text.New(pixel.V(55, 482), txt)
	others.Color = colornames.Red
	fmt.Fprintf(others, "Front")

	others.Dot = pixel.V(1035, 482)
	fmt.Fprintf(others, "Back")

	// Seats Columns
	seatNumsTop := text.New(pixel.V(104, 870), txt)
	seatNumsTop.Color = colornames.Black

	// Seat Rows
	seatRows := text.New(pixel.V(65, 827), txt)
	seatRows.Color = colornames.Darkblue

	labelString := ""
	charOrigin := pixel.V(105, 870)

	// Add labels for seat columns
	for i := 0; i < 24; i++ {
		if i <= 8 {
			charOrigin = pixel.V(float64(105+i*29), 870)
		} else {
			charOrigin = pixel.V(float64(100+i*29), 870)
		}

		seatNumsTop.Dot = charOrigin
		labelString = fmt.Sprintf("%d", i+1)
		fmt.Fprintf(seatNumsTop, labelString)
	}

	// Add labels for seat rows
	for i := 0; i < 6; i++ {
		if i < 3 {
			charOrigin = pixel.V(65, float64(827-i*17))
		} else {
			charOrigin = pixel.V(65, float64(811-i*17))
		}

		seatRows.Dot = charOrigin
		labelString = fmt.Sprintf("%c", 70-i)
		fmt.Fprintf(seatRows, labelString)
	}

	return seatNumsTop, seatRows, others
}

func run() {
	win := createWindow()
	plane := drawPlane()
	seatNumsTop, seatRows, others := drawLabels()

	generatePasses(WINDOW_TO_AISLE)
	//fmt.Println(passengers[0:6])

	board(6)

	// Passengers
	y := 503.
	pass := imdraw.New(nil)
	pass.Color = colornames.Limegreen

	// Entrance 1 starting point (30, 503)

	// Seats x = 52 + 60 * seatNumber

	for !win.Closed() {
		pass.Clear()
		win.Clear(colornames.Black)
		plane.Draw(win)

		others.Draw(win, pixel.IM.Scaled(others.Orig, 1.3))
		seatNumsTop.Draw(win, pixel.IM.Scaled(seatNumsTop.Orig, 1.8))
		seatRows.Draw(win, pixel.IM.Scaled(seatRows.Orig, 3))

		if y < 690 {
			y += 2
		}

		// Passengers
		pass.Push(pixel.V(30, y))
		pass.Circle(15, 0)
		pass.Draw(win)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
