package main

import "fmt"
import "os"
import "time"
import "encoding/json"
import "strings"

import "github.com/serg-2/libs-go/cryptolib"
import "github.com/serg-2/libs-go/loralib"
import "github.com/serg-2/libs-go/seriallib"

var message string
var receivedbytes byte
var send_message []byte
var message_source string
var key []byte
var x,y float64
var status bool

var conf Configuration

var send_signal,update_timer <-chan time.Time

type Configuration struct {
	Baud_rate   int
	Serial_port string
	Key         string
	Sending_timer int
        Update_coordinate_timer int
	Base_coordinates string
}

func update_coordinate() {
	for {
		<-update_timer
		for status !=true {
			x, y, status = seriallib.GetPosition("GGA", conf.Serial_port, conf.Baud_rate, true)
		}
		status = false
		//fmt.Println ("Coordinate updated")
	}
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
			message_source = fmt.Sprintf("%09.6f,%010.6f",x,y)
			send_message, _ = cryptolib.Encrypt(key, []byte(message_source))
			loralib.Send(send_message)
			fmt.Printf("Send: %s\n", message_source)

			loralib.ClearReceiver()

			loralib.ReceiveMode()

		default:
			status, received_message := loralib.CheckReceivedBuffer()
			if status {
				decrypted_message, _ := cryptolib.Decrypt(key, received_message)
				fmt.Printf("Payload: %s\n", string(decrypted_message))
				if len(strings.Split(string(decrypted_message),",")) != 2 {
					fmt.Println("err")
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func main() {

	file, err := os.Open("configuration.json")
	if err != nil {
		fmt.Printf("Can't open configuration file: configuration.json\n")
		os.Exit(0)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		fmt.Printf("Invalid JSON in configuration file.\n")
                os.Exit(0)
	}

	if len(os.Args[1:]) != 0 {
		fmt.Printf("Usage: %v\n", os.Args[0])
		os.Exit(0)
	}

	key = []byte(conf.Key) // 32 bytes \ 256 bit
	send_signal = time.Tick(time.Duration(conf.Sending_timer) * time.Second)
	update_timer = time.Tick(time.Duration(conf.Update_coordinate_timer) * time.Second)

	go update_coordinate()

	main_func()

}
