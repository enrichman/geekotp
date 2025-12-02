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

var display *ssd1306.Device

func main() {
	machine.InitSerial()
	fmt.Printf("--- GeekOTP Display Init ---\r\n")

	machine.I2C0.Configure(machine.I2CConfig{
		SDA:       machine.GPIO4,
		SCL:       machine.GPIO5,
		Frequency: 400_000,
	})

	display := ssd1306.NewI2C(machine.I2C0)
	display.Configure(ssd1306.Config{
		Address: 0x3C,
		Width:   128,
		Height:  32,
	})

	display.ClearDisplay()

	// Questo font è alto 8 pixel
	font := &proggy.TinySZ8pt7b
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	tinyfont.WriteLine(display, font, 0, 8, "GeekOTP Ready", white)

	display.Display()

	time.Sleep(3 * time.Second)
	display.ClearDisplay()

	// Loop to update the time without blinking
	for {
		// Erase the old time by drawing spaces in black
		// A larghezza di "10:00AM" è di 7 caratteri. Usiamo 8 spazi per sicurezza.

		// Draw the new time
		currentTime := time.Now().Format(time.TimeOnly)
		tinyfont.WriteLine(display, font, 0, 12, "           ", black)
		display.FillRectangle(0, 0, 128, 16, black)
		tinyfont.WriteLine(display, font, 0, 12, currentTime, white)

		// Update the display
		display.Display()

		// Print to console for debugging
		fmt.Print(currentTime, "\r\n")

		time.Sleep(time.Second)
	}
}
