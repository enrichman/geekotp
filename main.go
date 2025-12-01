package main

import (
	"fmt"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ssd1306"
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
		Height:  64,
	})

	display.ClearDisplay()
	///////////////////

	x := int16(0)
	y := int16(0)
	deltaX := int16(1)
	deltaY := int16(1)

	for {
		pixel := display.GetPixel(x, y)
		c := color.RGBA{255, 255, 255, 255}
		if pixel {
			c = color.RGBA{0, 0, 0, 255}
		}
		display.SetPixel(x, y, c)
		display.Display()

		x += deltaX
		y += deltaY

		if x == 0 || x == 127 {
			deltaX = -deltaX
		}
		if y == 0 || y == 63 {
			deltaY = -deltaY
		}
		time.Sleep(50 * time.Millisecond)
	}

}
