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
	IN_MENU_ITEM                 // State 2: Inside a selected menu item
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
	// Both pins are configured as Input with PULL-UP resistors (LOW when pressed).
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
			// --- OTP DISPLAY LOGIC (INIT STATE) ---
			lcd.SetCursor(0, 0)
			lcd.Write([]byte("**  GeekOTP   **"))
			lcd.Display()

			lcd.SetCursor(0, 1)
			lcd.Write([]byte("    v0.1    "))
			lcd.Display()
		case MENU:
			// --- MENU DISPLAY LOGIC (MENU STATE) ---
			if needsDisplayUpdate {
				updateSimpleMenuDisplay()
				needsDisplayUpdate = false // Reset the flag
			}
		}

		time.Sleep(time.Millisecond * 10)
	}
}

func handleInput() {
	// 1. Read Pin State
	isNavDown := !NAV_PIN.Get()
	isSelectDown := !SELECT_PIN.Get()

	// 2. Debounce Logic on Release (check elapsed time)
	if time.Since(lastInputTime) < time.Millisecond*150 {
		navPressed = isNavDown
		selectPressed = isSelectDown
		return
	}

	// 3. HANDLE NAVIGATION ACTION (NAV_PIN)
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
		case IN_MENU_ITEM:
			// No action for NAV in this state for now
		}
	}

	// 4. HANDLE SELECTION ACTION (SELECT_PIN)
	if isSelectDown {
		selectPressed = true
	} else if selectPressed {
		logger(fmt.Sprintf("SELECT pressed in state: %v", currentState))
		lastInputTime = time.Now()
		selectPressed = false
		needsDisplayUpdate = true

		switch currentState {
		case INIT:
			// Same as NAV, go to menu
			currentState = MENU
			currentMenuIndex = 0
		case MENU:
			// Select an item
			currentState = IN_MENU_ITEM
			selectedItem := menuOptions[currentMenuIndex]
			logger(fmt.Sprintf("Selected item: %s", selectedItem))

			// Execute action based on selection
			lcd.ClearDisplay()
			lcd.SetCursor(0, 0)
			if selectedItem == "code" {

				if time.Since(lastOtp) > time.Second {
					lastOtp = time.Now()
					now := time.Now()

					// Generate OTP code
					code, err := totp.GenerateCode(SECRET, now)
					if err != nil {
						logger("OTP generation error: " + err.Error())
					}

					code = fmt.Sprintf("%s-%s", code[:3], code[3:])

					// Calculate remaining time for the next code
					period := 30
					remainingSeconds := int64(period) - (now.Unix() % int64(period))

					// Update line 1 (OTP) only if the code has changed
					if code != lastOTPCode {
						lastOTPCode = code
						lcd.SetCursor(0, 0)
						lcd.Write([]byte("                ")) // Clear line
						lcd.SetCursor(0, 0)
						lcd.Write([]byte(code))
						lcd.Display()
					}

					// Update line 2 (countdown) only if the value has changed
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

			} else if selectedItem == "info" {
				lcd.Write([]byte("Info Page"))
				lcd.SetCursor(0, 1)
				lcd.Write([]byte("GeekOTP v0.1"))
				lcd.Display()
			}

		case IN_MENU_ITEM:
			// Go back to the menu
			currentState = MENU
		}
	}
}
