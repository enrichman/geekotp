package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/hd44780"
)

// === 1. DEFINIZIONI DEI PIN (GPIO) ===

// Pin LCD HD44780 (Interfaccia 4-bit)
const (
	// data pins
	D4_PIN = machine.GPIO8
	D5_PIN = machine.GPIO9
	D6_PIN = machine.GPIO10
	D7_PIN = machine.GPIO11

	// control pins
	RS_PIN = machine.GPIO6 // Register Select
	EN_PIN = machine.GPIO7 // Enable
)

// runDisplayLoop gestisce tutta la logica del display.
func runDisplayLoop() {
	println("--- Initializing Display ---")

	// 1. CONFIGURAZIONE LCD (HD44780)
	// RS, EN, D4, D5, D6, D7

	lcd, _ := hd44780.NewGPIO4Bit(
		[]machine.Pin{D4_PIN, D5_PIN, D6_PIN, D7_PIN},
		EN_PIN,
		RS_PIN,
		machine.NoPin,
	)

	lcd.Configure(hd44780.Config{
		Width:  16,
		Height: 2,
		// CursorOnOff: true,
		// CursorBlink: true,
	})

	lcd.Write([]byte("GeekOTP"))
	lcd.Display()

	time.Sleep(time.Second)

	for {
		currentTime := time.Now().Format(time.TimeOnly)

		lcd.SetCursor(0, 1)
		lcd.Write([]byte(currentTime))
		lcd.Display()

		// Print to console for debugging
		fmt.Print(currentTime, "\r\n")

		time.Sleep(time.Second)
	}
}
