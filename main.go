package main

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"log/slog"
	"machine"
	"time"

	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"

	"github.com/soypat/cyw43439"
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

	go func() {
		time.Sleep(10 * time.Second)
		fmt.Print("starting AP\r\n")

		dev := cyw43439.NewPicoWDevice()
		cfg := cyw43439.DefaultWifiConfig()
		cfg.Logger = slog.Default()

		err := dev.Init(cfg)
		logErr("Init", err)

		err = dev.StartAP("geekotp", "password", 1)
		logErr("StartAP", err)

		addr, err := dev.HardwareAddr6()
		logErr("HardwareAddr6", err)

		addrs := hex.EncodeToString(addr[:])
		println("MAC:", addrs[0], addrs[1], addrs[2], addrs[3], addrs[4], addrs[5])
	}()

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

func logErr(msg string, err error) {
	if err != nil {
		slog.Error(msg + ": " + err.Error())
	}
}
