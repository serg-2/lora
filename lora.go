package main

import "fmt"
import "os"
import "time"
import "strconv"

import "github.com/serg-2/go-aes-lib"
import "github.com/serg-2/go-lora-lib"

var message string
var receivedbytes byte
var send_message []byte
var message_source string
var key []byte

var send_signal <-chan time.Time

func main_func() {
	loralib.InitiateRPIO()
	loralib.SetupLoRa()
	loralib.ConfigSend()

	// Prepare to receive
	loralib.ReceiveMode()

	// Start MAIN CYCLE
	for {
		select {
		case <-send_signal:
			loralib.Send(send_message)
			fmt.Printf("Send: %s\n", message_source)

			loralib.ClearReceiver()

			loralib.ReceiveMode()

		default:
			status, received_message := loralib.CheckReceivedBuffer()
			if status {
				decrypted_message, _ := cryptolib.Decrypt(key, received_message)
				fmt.Printf("Payload: %s\n", string(decrypted_message))
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func main() {
	if len(os.Args[1:]) != 2 {
		fmt.Printf("Usage: %v <time to send> <string to send> \n", os.Args[0])
		os.Exit(0)
	}

	message_source = os.Args[2]
	time_from_arg, _ := strconv.Atoi(os.Args[1])

	key = []byte("Mega secret key SUper 123OGON !!") // 32 bytes \ 256 bit
	send_message, _ = cryptolib.Encrypt(key, []byte(message_source))

	send_signal = time.Tick(time.Duration(time_from_arg) * time.Second)

	main_func()

}
