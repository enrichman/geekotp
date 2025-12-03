package main

import (
	"fmt"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

var (
	display *ssd1306.Device

	font  = &proggy.TinySZ8pt7b
	white = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	black = color.RGBA{R: 0, G: 0, B: 0, A: 255}
)

func main() {
	// Attesa per dare il tempo al monitor seriale di connettersi.
	time.Sleep(2 * time.Second)
	machine.InitSerial()
	println("--- GeekOTP Starting ---")

	// Avviamo i loop del display e delle notifiche in goroutine separate.
	go runDisplayLoop()

	// Blocca il main. Il server BLE e le altre goroutine continueranno a girare.
	select {}
}

// runDisplayLoop gestisce tutta la logica del display.
func runDisplayLoop() {
	println("--- Initializing Display ---")
	machine.I2C0.Configure(machine.I2CConfig{SDA: machine.GPIO4, SCL: machine.GPIO5})
	display = ssd1306.NewI2C(machine.I2C0)

	display.Configure(ssd1306.Config{
		Width:  128,
		Height: 32,
	})

	display.ClearDisplay()
	tinyfont.WriteLine(display, font, 0, 8, "GeekOTP", white)
	tinyfont.WriteLine(display, font, 0, 16, "Status:", white)
	tinyfont.WriteLine(display, font, 0, 24, "> Waiting...", white)
	display.Display()

	for {
		currentTime := time.Now().Format(time.TimeOnly)
		display.FillRectangle(0, 0, 128, 16, black)
		tinyfont.WriteLine(display, font, 60, 8, currentTime, white)
		display.Display()

		// Print to console for debugging
		fmt.Print(currentTime, "\r\n")

		time.Sleep(time.Second)
	}
}
