package main

import (
	"fmt"
	"machine"
	"strconv"
	"time"

	"github.com/google/uuid"
	"tinygo.org/x/bluetooth"
)

// --- UUIDs per i nostri servizi e caratteristiche Bluetooth ---
var (
	serviceUUID, _    = uuid.Parse("7A11D3B2-B506-4E3A-A8E1-DD8B4A3F1C22")
	rwCharUUID, _     = uuid.Parse("A2BAE829-317A-443E-A03C-72E83149A52A")
	notifyCharUUID, _ = uuid.Parse("4F3551A5-A935-4228-B791-B7904B4239AC")
)

var (
	adapter       = bluetooth.DefaultAdapter
	currentDevice *bluetooth.Device
	temp          int32

	// Variabile per memorizzare l'ultimo valore ricevuto, così può essere letto.
	lastReceivedValue []byte

	// Handle per la caratteristica di notifica.
	notifyCharacteristic bluetooth.Characteristic
)

func bt() {
	go sendNotifications()
	go func() {
		for {
			temp = machine.ReadTemperature()
			time.Sleep(time.Second)
		}
	}()

	// --- Setup del Server Bluetooth ---
	must("enable BLE stack", adapter.Enable())

	adapter.SetConnectHandler(func(device bluetooth.Device, connected bool) {
		if connected {
			println("device connected:", device.Address.String())
			currentDevice = &device
		} else {
			println("device disconnected:", device.Address.String())
			currentDevice = nil
		}
	})

	// Aggiungiamo il nostro servizio custom con due caratteristiche.
	err := adapter.AddService(&bluetooth.Service{
		UUID: bluetooth.NewUUID(serviceUUID),
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  bluetooth.NewUUID(rwCharUUID),
				Flags: bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					fmt.Printf("BLE received: '%s'\r\n", string(value))
					lastReceivedValue = make([]byte, len(value))
					copy(lastReceivedValue, value)
				},
				Value: lastReceivedValue,
			},
			{
				Handle: &notifyCharacteristic,
				UUID:   bluetooth.NewUUID(notifyCharUUID),
				Flags:  bluetooth.CharacteristicNotifyPermission | bluetooth.CharacteristicReadPermission,
				Value:  []byte(strconv.Itoa(int(temp))),
			},
		},
	})
	must("add service", err)

	// Iniziamo a pubblicizzare il nostro servizio.
	adv := adapter.DefaultAdvertisement()
	must("configure advertisement", adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    "GeekOTP",
		ServiceUUIDs: []bluetooth.UUID{bluetooth.NewUUID(serviceUUID)},
	}))
	must("start advertising", adv.Start())
	println("Advertising BLE service...")

}

// sendNotifications invia un aggiornamento tramite BLE ogni 10 secondi.
func sendNotifications() {
	// Attende un po' per assicurarsi che il BLE sia pronto.
	time.Sleep(5 * time.Second)

	i := 0
	for {
		// Crea un messaggio da inviare.
		i++
		message := "Update #" + strconv.Itoa(i)

		// Invia il messaggio come notifica.
		// La libreria bluetooth si occupa di inviarlo solo ai client sottoscritti.

		n, err := notifyCharacteristic.Write([]byte(message))
		if err != nil {
			println("Error sending notification:", err.Error())
		} else {
			fmt.Printf("Sent notification: '%s', wrote %d bytes\r\n", message, n)
		}

		// Attende 10 secondi.
		time.Sleep(10 * time.Second)
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
