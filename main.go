package main

import (
	"fmt"
	"machine"
	"time"
)

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

func main() {
	// Attesa per dare il tempo al monitor seriale di connettersi.
	time.Sleep(2 * time.Second)

	machine.InitSerial()
	println("--- GeekOTP Starting ---")

	// 1. Configurazione dei Pin
	// Entrambi i pin sono configurati come Input con resistenza di PULL-UP (LOW quando premuto).
	NAV_PIN.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	SELECT_PIN.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	// Inizializza il timer Debounce
	lastInputTime = time.Now()

	// Avviamo i loop del display e delle notifiche in goroutine separate.
	go runDisplayLoop()

	// 2. Loop Principale
	for {
		handleInput()
		time.Sleep(time.Millisecond * 10)
	}
}

func handleInput() {
	// 1. Lettura Stato Pin
	isNavDown := !NAV_PIN.Get()
	isSelectDown := !SELECT_PIN.Get()

	// 2. Logica NAVIGAZIONE (GPIO 14)
	if isNavDown {
		// Il pulsante è premuto: lo registriamo
		navPressed = true
	} else {
		// Il pulsante è rilasciato
		if navPressed {
			// Se era stato premuto in precedenza, eseguiamo l'azione

			// Logica Debounce sul Rilascio:
			if time.Since(lastInputTime) >= time.Millisecond*150 {
				fmt.Printf("-> AZIONE NAVIGAZIONE (RILASCIO)\r\n")
				// [QUI ANDRÀ LA LOGICA DI SCORRIMENTO DEL MENU]
				lastInputTime = time.Now()
			}

			// Resettiamo lo stato di pressione
			navPressed = false
		}
	}

	// 3. Logica SELEZIONE (GPIO 15)
	if isSelectDown {
		// Il pulsante è premuto: lo registriamo
		selectPressed = true
	} else {
		// Il pulsante è rilasciato
		if selectPressed {
			// Se era stato premuto in precedenza, eseguiamo l'azione

			// Logica Debounce sul Rilascio:
			if time.Since(lastInputTime) >= time.Millisecond*150 {
				fmt.Printf("-> AZIONE SELECT (RILASCIO)\r\n")
				// [QUI ANDRÀ LA LOGICA DI SELEZIONE DEL MENU]
				lastInputTime = time.Now()
			}

			// Resettiamo lo stato di pressione
			selectPressed = false
		}
	}
}
