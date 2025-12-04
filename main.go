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

// === DEFINIZIONI PIN HARDWARE (2 PULSANTI) ===

// Pulsante 1: Navigazione (NAV)
const NAV_PIN = machine.GPIO14

// Pulsante 2: Selezione (SEL)
const SELECT_PIN = machine.GPIO15

// Variabili per tracciare lo stato di pressione dei pulsanti
var navPressed bool
var selectPressed bool

// Variabile per il Debounce (ignora le letture troppo veloci)
var lastInputTime time.Time

type AppState int

const (
	INIT AppState = iota // Stato 0: Schermata iniziale/bloccata
	MENU                 // Stato 1: Navigazione Menu
)

type MenuCategory struct {
	Title   string       // Riga 1: Nome della Categoria
	Options []MenuOption // Riga 2: Le opzioni all'interno
}

type MenuOption struct {
	Text   string
	Action func()
}

// Struttura del menu gerarchico
var menu = []MenuCategory{
	{
		Title: "Bluetooth:",
		Options: []MenuOption{
			{"On", func() { logger("Bluetooth ON") }},
			{"Off", func() { logger("Bluetooth OFF") }},
			{"<- Back", nil},
		},
	},
	{
		Title: "Info:",
		Options: []MenuOption{
			{"Version", func() { logger("Mostra Versione") }},
			{"Serial", func() { logger("Mostra Seriale") }},
			{"<- Back", nil},
		},
	},
}

const SECRET = ""

var (
	currentState AppState = INIT

	// Stato Menu
	currentCategoryIndex int = 0
	currentOptionIndex   int = 0

	needsDisplayUpdate bool
)

var lastOtp time.Time
var lastOTPCode string
var lastRemainingSeconds int64

func main() {
	lastOtp = time.Now()

	// Attesa per dare il tempo al monitor seriale di connettersi.
	time.Sleep(2 * time.Second)

	machine.InitSerial()
	logger("--- GeekOTP Starting ---")

	// 1. Configurazione dei Pin
	// Entrambi i pin sono configurati come Input con resistenza di PULL-UP (LOW quando premuto).
	NAV_PIN.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	SELECT_PIN.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	// Inizializza il timer Debounce
	lastInputTime = time.Now()

	initDisplay()
	lcd.ClearDisplay() // Pulisce lo schermo una sola volta all'avvio

	// 2. Loop Principale
	for {
		handleInput()

		// --- LOGICA DI VISUALIZZAZIONE OTP (STATO INIT) ---
		if time.Since(lastOtp) > time.Second {
			lastOtp = time.Now()
			now := time.Now()

			// Genera il codice OTP
			code, err := totp.GenerateCode(SECRET, now)
			if err != nil {
				logger("Errore generazione OTP: " + err.Error())
				continue
			}

			code = fmt.Sprintf("%s-%s", code[:3], code[3:])

			// Calcola il tempo rimanente per il prossimo codice
			period := 30
			remainingSeconds := int64(period) - (now.Unix() % int64(period))

			// Aggiorna la riga 1 (OTP) solo se il codice è cambiato
			if code != lastOTPCode {
				lastOTPCode = code
				lcd.SetCursor(0, 0)
				lcd.Write([]byte(code))
				lcd.Display()
			}

			// Aggiorna la riga 2 (countdown) solo se il valore è cambiato
			if remainingSeconds != lastRemainingSeconds {
				lastRemainingSeconds = remainingSeconds
				lcd.SetCursor(0, 1)
				lcd.Write([]byte("                ")) // Pulisce la riga 2
				lcd.SetCursor(0, 1)
				line2 := fmt.Sprintf("Next in: %2ds", remainingSeconds)
				lcd.Write([]byte(line2))
				lcd.Display()
			}
		}

		// --- LOGICA DI VISUALIZZAZIONE MENU (STATO MENU) ---
		// if needsDisplayUpdate {
		// 	updateMenuDisplay()
		// 	needsDisplayUpdate = false // Resetta il flag
		// }

		time.Sleep(time.Millisecond * 10)
	}
}

func handleInput() {
	// 1. Lettura Stato Pin
	isNavDown := !NAV_PIN.Get()
	isSelectDown := !SELECT_PIN.Get()

	// 2. Logica Debounce sul Rilascio (controlla il tempo trascorso)
	if time.Since(lastInputTime) < time.Millisecond*150 {
		// Aggiorna solo gli stati di pressione, ma esce
		navPressed = isNavDown
		selectPressed = isSelectDown
		return
	}

	// 3. ESECUZIONE AZIONI NAVIGAZIONE (NAV_PIN)
	if isNavDown {
		navPressed = true // Registra la pressione
	} else if navPressed {
		logger(fmt.Sprintf("IN currentState: %v - navpressed", currentState))

		lastInputTime = time.Now()
		navPressed = false // Resetta lo stato di pressione
		needsDisplayUpdate = true

		switch currentState {
		case INIT:
			currentState = MENU
			currentCategoryIndex = 0
		case MENU:
			currentCategoryIndex++
			if currentCategoryIndex >= len(menu) {
				currentCategoryIndex = 0
			}
		}
	}

	// 4. ESECUZIONE AZIONI SELEZIONE (SELECT_PIN)
	if isSelectDown {
		selectPressed = true // Registra la pressione
	} else if selectPressed {
		logger(fmt.Sprintf("IN currentState: %v - selectPressed", currentState))

		lastInputTime = time.Now()
		selectPressed = false // Resetta lo stato di pressione
		needsDisplayUpdate = true

		switch currentState {
		case INIT:
			currentState = MENU
			currentCategoryIndex = 0
		case MENU:
			currentCategoryIndex++
			if currentCategoryIndex >= len(menu) {
				currentCategoryIndex = 0
			}
		}
	}
}
