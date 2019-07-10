package main

import "fmt"
import "os"
import "time"
import "strconv"
import "encoding/json"

import "github.com/serg-2/libs-go/cryptolib"
import "github.com/serg-2/libs-go/loralib"
import "github.com/serg-2/libs-go/seriallib"

var message string
var receivedbytes byte
var send_message []byte
var message_source string
var key []byte

var send_signal <-chan time.Time

type Configuration struct {
	Baud_rate   int
	Serial_port string
	Key         string
}

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
	var conf Configuration
	//filename is the path to the json config file
	file, err := os.Open("configuration.json")
	if err != nil {
		return
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		return
	}

	if len(os.Args[1:]) != 2 {
		fmt.Printf("Usage: %v <time to send> <string to send> \n", os.Args[0])
		os.Exit(0)
	}

	message_source = os.Args[2]
	time_from_arg, _ := strconv.Atoi(os.Args[1])

	key = []byte(conf.Key) // 32 bytes \ 256 bit
	send_message, _ = cryptolib.Encrypt(key, []byte(message_source))

	send_signal = time.Tick(time.Duration(time_from_arg) * time.Second)

	//Serial Part
	x, y, status := seriallib.GetPosition("GGA", conf.Serial_port, conf.Baud_rate, true)
	if status {
		fmt.Printf("%v --- %v\n", x, y)
	}

	main_func()

}
