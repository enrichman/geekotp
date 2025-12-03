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
		// Controllo del Debounce: 150ms
		if time.Since(lastInputTime) >= time.Millisecond*150 {

			// Legge lo stato del pin: !Get() è TRUE se premuto.
			if !NAV_PIN.Get() {
				fmt.Print("-> PULSANTE NAVIGAZIONE PREMUTO (GPIO 14)\r\n")
				lastInputTime = time.Now()
			} else if !SELECT_PIN.Get() {
				fmt.Print("-> PULSANTE SELECT PREMUTO (GPIO 15)\r\n")
				lastInputTime = time.Now()
			}
		}

		// Pausa essenziale per lo scheduler cooperativo di TinyGo
		time.Sleep(time.Millisecond * 10)
	}
}
