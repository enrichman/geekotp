package main

import (
	"fmt"
	"machine"
	"time"

	"github.com/pquerna/otp/totp"
)

func logger(msg string) {
	fmt.Printf(msg + "\r\n")
}

// === HARDWARE PIN DEFINITIONS (2 BUTTONS) ===

// Button 1: Navigation (NAV)
const NAV_PIN = machine.GPIO14

// Button 2: Selection (SEL)
const SELECT_PIN = machine.GPIO15

// Variables to track button press states
var navPressed bool
var selectPressed bool

// Debounce variable (ignores rapid inputs)
var lastInputTime time.Time

// Application state definitions
type AppState int

const (
	INIT         AppState = iota // State 0: Initial/OTP screen
	MENU                         // State 1: Menu navigation
	IN_MENU_ITEM                 // State 2: Inside a selected menu item (for static pages like 'info')
	SHOW_CODE                    // State 3: Displaying the OTP code
)

// Simple menu structure
var menuOptions = []string{
	"code",
	"info",
}

var (
	currentState       AppState = INIT
	currentMenuIndex   int      = 0
	needsDisplayUpdate bool
)

const SECRET = "JBSWY3DPEHPK3PXP"

var lastOtp time.Time
var lastOTPCode string
var lastRemainingSeconds int64

func main() {
	lastOtp = time.Now()

	// Wait for the serial monitor to connect
	time.Sleep(2 * time.Second)

	machine.InitSerial()
	logger("--- GeekOTP Starting ---")

	// 1. Configure Pins
	NAV_PIN.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	SELECT_PIN.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	// Initialize Debounce timer
	lastInputTime = time.Now()

	initDisplay()
	lcd.ClearDisplay() // Clear screen once on startup

	// 2. Main Loop
	for {
		handleInput()

		switch currentState {
		case INIT:
			lcd.SetCursor(0, 0)
			lcd.Write([]byte("**  GeekOTP   **"))
			lcd.Display()
			lcd.SetCursor(0, 1)
			lcd.Write([]byte("    v0.1    "))
			lcd.Display()
		case MENU:
			if needsDisplayUpdate {
				updateSimpleMenuDisplay()
				needsDisplayUpdate = false // Reset the flag
			}
		case SHOW_CODE:
			updateOTP()
		}

		time.Sleep(time.Millisecond * 10)
	}
}

func updateOTP() {
	if time.Since(lastOtp) > time.Second {
		lastOtp = time.Now()
		now := time.Now()

		code, err := totp.GenerateCode(SECRET, now)
		if err != nil {
			logger("OTP generation error: " + err.Error())
		}
		code = fmt.Sprintf("%s-%s", code[:3], code[3:])

		period := 30
		remainingSeconds := int64(period) - (now.Unix() % int64(period))

		if code != lastOTPCode {
			lastOTPCode = code
			lcd.SetCursor(0, 0)
			lcd.Write([]byte("                ")) // Clear line
			lcd.SetCursor(0, 0)
			lcd.Write([]byte(code))
			lcd.Display()
		}

		if remainingSeconds != lastRemainingSeconds {
			lastRemainingSeconds = remainingSeconds
			lcd.SetCursor(0, 1)
			lcd.Write([]byte("                ")) // Clear line
			lcd.SetCursor(0, 1)
			line2 := fmt.Sprintf("Next in: %2ds", remainingSeconds)
			lcd.Write([]byte(line2))
			lcd.Display()
		}
	}
}

func handleInput() {
	isNavDown := !NAV_PIN.Get()
	isSelectDown := !SELECT_PIN.Get()

	if time.Since(lastInputTime) < time.Millisecond*150 {
		navPressed = isNavDown
		selectPressed = isSelectDown
		return
	}

	if isNavDown {
		navPressed = true
	} else if navPressed {
		logger(fmt.Sprintf("NAV pressed in state: %v", currentState))
		lastInputTime = time.Now()
		navPressed = false
		needsDisplayUpdate = true

		switch currentState {
		case INIT:
			currentState = MENU
			currentMenuIndex = 0
		case MENU:
			currentMenuIndex++
			if currentMenuIndex >= len(menuOptions) {
				currentMenuIndex = 0 // Loop back
			}
		case IN_MENU_ITEM, SHOW_CODE:
			// No action for NAV in these states
		}
	}

	if isSelectDown {
		selectPressed = true
	} else if selectPressed {
		logger(fmt.Sprintf("SELECT pressed in state: %v", currentState))
		lastInputTime = time.Now()
		selectPressed = false
		needsDisplayUpdate = true

		switch currentState {
		case INIT:
			currentState = MENU
			currentMenuIndex = 0
		case MENU:
			selectedItem := menuOptions[currentMenuIndex]
			logger(fmt.Sprintf("Selected item: %s", selectedItem))

			lcd.ClearDisplay()
			if selectedItem == "code" {
				currentState = SHOW_CODE
			} else if selectedItem == "info" {
				currentState = IN_MENU_ITEM
				lcd.SetCursor(0, 0)
				lcd.Write([]byte("Info Page"))
				lcd.SetCursor(0, 1)
				lcd.Write([]byte("GeekOTP v0.1"))
				lcd.Display()
			}

		case IN_MENU_ITEM, SHOW_CODE:
			// Go back to the menu from 'info' or 'code' screen
			currentState = MENU
			lastOTPCode = ""
		}
	}
}
