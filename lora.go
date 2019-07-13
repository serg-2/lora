package main

import "fmt"
import "os"
import "time"
import "encoding/json"
import "strings"
import "strconv"

import "github.com/serg-2/libs-go/cryptolib"
import "github.com/serg-2/libs-go/loralib"
import "github.com/serg-2/libs-go/seriallib"
import "github.com/serg-2/libs-go/marinelib"

var message string
var receivedbytes byte
var send_message []byte
var message_source string
var key []byte
var myposition, recposition, baseposition [2]float64
var status bool

var conf Configuration

var send_signal_frequency, local_update_timer <-chan time.Time

type Configuration struct {
	Baud_rate        int
	Serial_port      string
	Key              string
	Base_coordinates string
	Running_mode     string
}

func update_coordinate() {
	for {
		<-local_update_timer
		for status != true {
			myposition, status = seriallib.GetPosition("GGA", conf.Serial_port, conf.Baud_rate, true)
			status = false
			//fmt.Println("Coordinate updated")
		}
	}
}

func parsefloat(s string) float64 {
	_r, _ := strconv.ParseFloat(s, 64)
	return _r
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
		case <-send_signal_frequency:
			message_source = fmt.Sprintf("%09.6f,%010.6f", myposition[0], myposition[1])
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
				if len(strings.Split(string(decrypted_message), ",")) == 2 {
					recposition[0] = parsefloat(strings.Split(string(decrypted_message), ",")[0])
					recposition[1] = parsefloat(strings.Split(string(decrypted_message), ",")[1])
					fmt.Printf("Distance: %s\n", fmt.Sprintf("%5.2f", marinelib.CalculateDistance(recposition, baseposition)))
					fmt.Printf("Bearing: %s\n", fmt.Sprintf("%5.1f", marinelib.CalculateBearing(recposition, baseposition)))
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

	baseposition = [2]float64{parsefloat(strings.Split(conf.Base_coordinates, ",")[0]), parsefloat(strings.Split(conf.Base_coordinates, ",")[1])}

	if conf.Running_mode == "base_station" {
		send_signal_frequency = time.Tick(time.Duration(3) * time.Second)
		local_update_timer = time.Tick(time.Duration(-1) * time.Second)
		//DEBUG
		myposition = baseposition
	} else {
		send_signal_frequency = time.Tick(time.Duration(500) * time.Second)
		local_update_timer = time.Tick(time.Duration(1) * time.Second)
	}

	go update_coordinate()

	main_func()

}
