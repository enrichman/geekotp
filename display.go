package main

import (
	"fmt"
	"machine"

	"tinygo.org/x/drivers/hd44780"
)

// === 1. PIN DEFINITIONS (GPIO) ===

// HD44780 LCD Pins (4-bit interface)
const (
	// data pins
	D4_PIN = machine.GPIO8
	D5_PIN = machine.GPIO9
	D6_PIN = machine.GPIO10
	D7_PIN = machine.GPIO11

	// control pins
	RS_PIN = machine.GPIO6 // Register Select
	EN_PIN = machine.GPIO7 // Enable

	// backlight pin
	BACKLIGHT_PIN = machine.GPIO16
)

var lcd hd44780.Device

// initDisplay initializes and configures the LCD.
func initDisplay() {
	logger("--- Initializing Display ---")

	// Backlight
	BACKLIGHT_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	BACKLIGHT_PIN.High()

	var err error
	lcd, err = hd44780.NewGPIO4Bit(
		[]machine.Pin{D4_PIN, D5_PIN, D6_PIN, D7_PIN},
		EN_PIN,
		RS_PIN,
		machine.NoPin,
	)
	if err != nil {
		logger("Failed to create new GPIO4Bit device: " + err.Error())
		return
	}

	config := hd44780.Config{
		Width:  16,
		Height: 2,
	}

	if err := lcd.Configure(config); err != nil {
		logger("Failed to configure display: " + err.Error())
		return
	}

	lcd.Write([]byte("GeekOTP Starting"))
	lcd.Display()
}

// updateSimpleMenuDisplay updates the display for the simple menu.
func updateSimpleMenuDisplay() {
	logger("Updating simple menu display...")

	lcd.ClearDisplay()

	// Line 1: Menu Title
	lcd.SetCursor(0, 0)
	lcd.Write([]byte("Menu:"))
	lcd.Display()

	// Line 2: Selected Option with ">"
	optionText := menuOptions[currentMenuIndex]
	displayText := fmt.Sprintf("> %s", optionText)

	lcd.SetCursor(0, 1)
	lcd.Write([]byte(displayText))
	lcd.Display()

	logger("Display updated with: " + displayText)
}

func turnScreenOn() {
	logger("Turning screen on")
	BACKLIGHT_PIN.High()
}

func turnScreenOff() {
	logger("Turning screen off")
	BACKLIGHT_PIN.Low()
}
