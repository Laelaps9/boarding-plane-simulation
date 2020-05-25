package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
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
const HANDLING_BAGS = 1
const SITTING = 2

// Order masking
const RANDOM = 0
const BACK_TO_FRONT = 1
const FRONT_TO_BACK = 2
const WINDOW_TO_AISLE = 3
const AISLE_TO_WINDOW = 4

// Global varaibles
var elapsed = 0
var seated = 0
var passengers []Passenger

type Passenger struct {
	PosX     int
	PosY     int
	SeatN    int
	SeatL    string
	State    int
	Delay    int
	BagsDone bool
}

func copySlice(dest []Passenger, orig []Passenger) []Passenger {
	for i := range dest {
		dest[i] = orig[i]
	}
	return dest
}

func generatePasses(size int, orderFlag int) {

	var seats [144]int
	rows := [6]string{"A", "B", "C", "D", "E", "F"}

	// Create boarding passes
	for i := range seats {
		seats[i] = i
	}

	var j = 0
	var prePassengers = make([]Passenger, 144)

	// Assign boarding passes
	for i := range prePassengers {

		if i%24 == 0 && i != 0 {
			j++
		}

		prePassengers[i].PosX = 0
		prePassengers[i].PosY = 0

		prePassengers[i].SeatN = (i % 24) + 1
		prePassengers[i].SeatL = rows[j]

		prePassengers[i].State = STANDING
		prePassengers[i].Delay = 0
		prePassengers[i].BagsDone = false
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(prePassengers), func(i, j int) { prePassengers[i], prePassengers[j] = prePassengers[j], prePassengers[i] })

	passengers = make([]Passenger, size)
	passengers = copySlice(passengers, prePassengers)

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

	} else if orderFlag == AISLE_TO_WINDOW {
		tmpPassengers := make([]Passenger, 144)
		var i = 0

		for j := range passengers {
			if passengers[j].SeatL == "C" || passengers[j].SeatL == "D" {
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
			if passengers[j].SeatL == "A" || passengers[j].SeatL == "F" {
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

func getPassengerInPosition(PosX int, PosY int) int {
	for i := range passengers {
		if passengers[i].PosX == PosX && passengers[i].PosY == PosY {
			return i
		}
	}
	return -1
}

func swapPassengers(walker int, obstructer int) {
	passengers[walker].Delay += 4
	passengers[obstructer].Delay += 4
}

func board(size int, win *pixelgl.Window, plane, draw *imdraw.IMDraw, seatNumsTop, seatRows, others, results *text.Text) {

	for {

		// Uncomment to print behaviour chains at each iteration
		// fmt.Println(passengers)

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
					passengers[i].State = HANDLING_BAGS
					passengers[i].Delay = 18
				} else if passengers[i].Delay != 0 && passengers[i].State == HANDLING_BAGS {
					passengers[i].Delay--
				} else if passengers[i].Delay == 0 && passengers[i].State == HANDLING_BAGS {
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
							swapPassengers(i, obstructer)
							passengers[i].PosY--
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
							swapPassengers(i, obstructer)
							passengers[i].PosY--
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
							swapPassengers(i, obstructer)
							passengers[i].PosY--
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
							swapPassengers(i, obstructer)
							passengers[i].PosY++
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
							swapPassengers(i, obstructer)
							passengers[i].PosY++
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
							swapPassengers(i, obstructer)
							passengers[i].PosY++
						}

						// Sit down when seat if found
					} else if passengers[i].SeatL == "F" && passengers[i].PosY == 3 && passengers[i].State != SITTING {
						passengers[i].State = SITTING
						seated++
					}
				}

			}

		}

		printDrawings(win, plane, seatNumsTop, seatRows, others, results)
		draw.Clear()
		drawPassengers(passengers[0:size], win, draw)

		elapsed++

	}

	var swapDelays = 0

	for i := range passengers {
		swapDelays += passengers[i].Delay
	}

	elapsed += swapDelays
	//elapsed *= 2

	//fmt.Println("Swap Delays:", swapDelays, "s")
	fmt.Println("Elapsed:", elapsed/60, "min")
	fmt.Fprintf(results, "Total Passengers: %d \t Elapsed time: %d min", size, (elapsed/60))
	printDrawings(win, plane, seatNumsTop, seatRows, others, results)
	draw.Clear()
	drawPassengers(passengers[0:size], win, draw)

}

func createWindow() *pixelgl.Window {
	// Specify configuration window
	cfg := pixelgl.WindowConfig{
		Title:  "Plane Boarding Simulator",
		Bounds: pixel.R(0, 400, 1430, 900),
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

func drawLabels() (*text.Text, *text.Text, *text.Text, *text.Text) {
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

	// Results
	results := text.New(pixel.V(715, 450), txt)
	results.Color = colornames.White

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

	return seatNumsTop, seatRows, others, results
}

func drawPassengers(passengers []Passenger, win *pixelgl.Window, draw *imdraw.IMDraw) {
	for i := range passengers {
		if passengers[i].PosX >= 0 {
			draw.Push(pixel.V(float64(52*(passengers[i].PosX)+60), float64(690+passengers[i].PosY*50)))
			draw.Circle(20, 0)
		}
	}
	draw.Draw(win)
	win.Update()
	//time.Sleep(time.Second / 2)
}

func printDrawings(win *pixelgl.Window, plane *imdraw.IMDraw, seatNumsTop, seatRows, others, results *text.Text) {
	win.Clear(colornames.Black)
	plane.Draw(win)
	seatNumsTop.Draw(win, pixel.IM.Scaled(seatNumsTop.Orig, 1.8))
	seatRows.Draw(win, pixel.IM.Scaled(seatRows.Orig, 3))
	others.Draw(win, pixel.IM.Scaled(others.Orig, 1.3))
	results.Draw(win, pixel.IM.Scaled(results.Orig, 2))
}

func run() {

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter number of passengers: [1-144]")
	sizeInput, _ := reader.ReadString('\n')
	sizeInput = strings.Replace(sizeInput, "\n", "", -1)
	size, _ := strconv.Atoi(sizeInput)

	fmt.Println("Select a bording method:\n [0] Random\n [1] Back to front\n [2] Front to back\n [3] Window to aisle\n [4] Aisle to window")
	orderInput, _ := reader.ReadString('\n')
	orderInput = strings.Replace(orderInput, "\n", "", -1)
	order, _ := strconv.Atoi(orderInput)

	win := createWindow()
	plane := drawPlane()
	seatNumsTop, seatRows, others, results := drawLabels()

	generatePasses(size, order)

	// Passengers
	pass := imdraw.New(nil)
	pass.Color = colornames.Limegreen

	board(size, win, plane, pass, seatNumsTop, seatRows, others, results)
	for !win.Closed() {

	}
}

func main() {
	pixelgl.Run(run)
}
