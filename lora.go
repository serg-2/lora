package main

import "fmt"
import "os"
import "time"
import "strconv"

import "github.com/stianeikeland/go-rpio"
import "log"

//Vars for Pi WiringPI
//var ssPin int = 6
//var dio0 int = 7
//var RST int = 0

//Vars for go-rpio
var ssPin rpio.Pin = 25
var dio0 rpio.Pin = 4
var RST rpio.Pin = 17

var CHANNEL uint8 = 0

const go_REG_OPMODE = 0x01
const go_OPMODE_MASK = 0x07
const go_not_OPMODE_MASK = 0xF8
const go_OPMODE_SLEEP = 0x00
const go_freq = 868100000
const go_REG_FRF_MSB = 0x06
const go_REG_FRF_MID = 0x07
const go_REG_FRF_LSB = 0x08
const go_REG_SYNC_WORD = 0x39
const go_REG_MODEM_CONFIG = 0x1D
const go_REG_MODEM_CONFIG2 = 0x1E
const go_REG_MODEM_CONFIG3 = 0x26
const go_REG_SYMB_TIMEOUT_LSB = 0x1F

const go_REG_VERSION = 0x42

const go_REG_MAX_PAYLOAD_LENGTH = 0x23
const go_REG_PAYLOAD_LENGTH = 0x22
const go_PAYLOAD_LENGTH = 0x40
const go_REG_HOP_PERIOD = 0x24
const go_REG_FIFO_ADDR_PTR = 0x0D

const go_REG_FIFO_RX_BASE_AD = 0x0F
const go_REG_LNA = 0x0C
const go_LNA_MAX_GAIN = 0x23
const go_REG_FIFO = 0x00

const go_OPMODE_LORA = 0x80
const go_OPMODE_STANDBY = 0x01
const go_OPMODE_RX = 0x05
const go_OPMODE_TX = 0x03

const go_RegPaRamp = 0x0A
const go_RegPaConfig = 0x09
const go_RegPaDac = 0x5A

const go_RegDioMapping1 = 0x40
const go_MAP_DIO0_LORA_TXDONE = 0x40
const go_MAP_DIO1_LORA_NOP = 0x30
const go_MAP_DIO2_LORA_NOP = 0xC0
const go_REG_IRQ_FLAGS = 0x12
const go_REG_IRQ_FLAGS_MASK = 0x11
const go_IRQ_LORA_TXDONE_MASK = 0x08
const go_not_IRQ_LORA_TXDONE_MASK = 0xF7
const go_REG_FIFO_TX_BASE_AD = 0x0E

const go_REG_PKT_SNR_VALUE = 0x19
const go_REG_FIFO_RX_CURRENT_ADDR = 0x10
const go_REG_RX_NB_BYTES = 0x13

const (
	SF7  = 7
	SF8  = 8
	SF9  = 9
	SF10 = 10
	SF11 = 11
	SF12 = 12
)

const go_sf = SF7

var sx1272 bool

var message string
var receivedbytes byte
var send_message string

var send_signal <-chan time.Time

func go_selectreceiver() {
        rpio.WritePin(ssPin, rpio.Low)
}

func go_unselectreceiver() {
        rpio.WritePin(ssPin, rpio.High)
}

func go_writeReg(addr byte, value byte) {
        var spibuf byte = addr | 0x80
        go_selectreceiver()
        rpio.SpiTransmit(spibuf,value)
        go_unselectreceiver()
}

func go_readReg(addr byte) byte {
        var spibuf []byte
        spibuf = make([]byte, 2)
        go_selectreceiver()
        spibuf[0] = addr & 0x7F
        spibuf[1] = 0x00
        rpio.SpiExchange(spibuf)     
        go_unselectreceiver()
        return byte(spibuf[1])
}


func go_opmode(mode byte) {
	go_writeReg(go_REG_OPMODE, go_readReg(go_REG_OPMODE)&go_not_OPMODE_MASK|mode)
}

func go_opmodeLora() {
	var u byte = go_OPMODE_LORA
	if sx1272 == false {
		u |= 0x8 // TBD: sx1276 high freq
	}
	go_writeReg(go_REG_OPMODE, u)
}

func go_SetupLoRa() {
        rpio.WritePin(RST, rpio.High)
	time.Sleep(100 * time.Millisecond)
        rpio.WritePin(RST, rpio.Low)
	time.Sleep(100 * time.Millisecond)
	var version byte = go_readReg(go_REG_VERSION)
	if version == 0x22 {
		// sx1272
		fmt.Println("SX1272 detected, starting.")
		sx1272 = true
	} else {
		// sx1276?
                rpio.WritePin(RST, rpio.Low)
		time.Sleep(100 * time.Millisecond)
                rpio.WritePin(RST, rpio.High)
		time.Sleep(100 * time.Millisecond)
		version = go_readReg(go_REG_VERSION)
		if version == 0x12 {
			// sx1276
			fmt.Println("SX1276 detected, Starting.")
			sx1272 = false
		} else {
			fmt.Println("Unrecognized transceiver.")
			//fmt.Printf("Transceiver version %x",version)
			os.Exit(1)
		}
	}

	go_opmode(go_OPMODE_SLEEP)

	//set frequency
	var frf uint64 = uint64(go_freq<<19) / 32000000
	go_writeReg(go_REG_FRF_MSB, byte(frf>>16))
	go_writeReg(go_REG_FRF_MID, byte(frf>>8))
	go_writeReg(go_REG_FRF_LSB, byte(frf>>0))

	go_writeReg(go_REG_SYNC_WORD, 0x34) //LoRaWAN public sync word

	if sx1272 {
		if go_sf == SF11 || go_sf == SF12 {
			go_writeReg(go_REG_MODEM_CONFIG, 0x0B)
		} else {
			go_writeReg(go_REG_MODEM_CONFIG, 0x0A)
		}
		go_writeReg(go_REG_MODEM_CONFIG2, (go_sf<<4)|0x04)
	} else {
		if go_sf == SF11 || go_sf == SF12 {
			go_writeReg(go_REG_MODEM_CONFIG3, 0x0C)
		} else {
			go_writeReg(go_REG_MODEM_CONFIG3, 0x04)
		}
		go_writeReg(go_REG_MODEM_CONFIG, 0x72)
		go_writeReg(go_REG_MODEM_CONFIG2, (go_sf<<4)|0x04)
	}

	if go_sf == SF10 || go_sf == SF11 || go_sf == SF12 {
		go_writeReg(go_REG_SYMB_TIMEOUT_LSB, 0x05)
	} else {
		go_writeReg(go_REG_SYMB_TIMEOUT_LSB, 0x08)
	}
	go_writeReg(go_REG_MAX_PAYLOAD_LENGTH, 0x80)
	go_writeReg(go_REG_PAYLOAD_LENGTH, go_PAYLOAD_LENGTH)
	go_writeReg(go_REG_HOP_PERIOD, 0xFF)
	go_writeReg(go_REG_FIFO_ADDR_PTR, go_readReg(go_REG_FIFO_RX_BASE_AD))

	go_writeReg(go_REG_LNA, go_LNA_MAX_GAIN)
}

func go_configPower(pw int8) {
	if sx1272 == false {
		// no boost used for now
		if pw >= 17 {
			pw = 15
		} else if pw < 2 {
			pw = 2
		}
		// check board type for BOOST pin
		go_writeReg(go_RegPaConfig, byte(0x80|byte(pw&0xf)))
		go_writeReg(go_RegPaDac, go_readReg(go_RegPaDac)|0x4)
	} else {
		// set PA config (2-17 dBm using PA_BOOST)
		if pw > 17 {
			pw = 17
		} else if pw < 2 {
			pw = 2
		}
		go_writeReg(go_RegPaConfig, byte(0x80|byte(pw-2)))
	}
}

func go_txlora(send_string string) {
	// set the IRQ mapping DIO0=TxDone DIO1=NOP DIO2=NOP
	go_writeReg(go_RegDioMapping1, go_MAP_DIO0_LORA_TXDONE|go_MAP_DIO1_LORA_NOP|go_MAP_DIO2_LORA_NOP)
	// clear all radio IRQ flags
	go_writeReg(go_REG_IRQ_FLAGS, 0xFF)
	// mask all IRQs but TxDone
	go_writeReg(go_REG_IRQ_FLAGS_MASK, go_not_IRQ_LORA_TXDONE_MASK)

	// initialize the payload size and address pointers
	go_writeReg(go_REG_FIFO_TX_BASE_AD, 0x00)
	go_writeReg(go_REG_FIFO_ADDR_PTR, 0x00)
	go_writeReg(go_REG_PAYLOAD_LENGTH, byte(len(send_string)))

	// download buffer to the radio FIFO
        writeBuf(go_REG_FIFO, send_string)
	// now we actually start the transmission
	go_opmode(go_OPMODE_TX)

	fmt.Printf("send: %s\n", send_string)

}

func writeBuf(addr byte, send_string string) {
        string_by_bytes := make([]byte, 1+len(send_string))
        string_by_bytes[0] = addr | 0x80
        for i := range send_string {
            string_by_bytes[i+1] = []byte(send_string)[i]
        }
        go_selectreceiver()
        rpio.SpiTransmit(string_by_bytes...)
        go_unselectreceiver()
}

func go_receivepacket() {
	var SNR int
	var rssicorr int

        if int(rpio.ReadPin(dio0)) == 1 {
		if a, message := go_receive(); a {
			value := byte(go_readReg(go_REG_PKT_SNR_VALUE))
			if (value & 0x80) == 1 { //The SNR sign bit is 1
				// Invert and divide by 4
				value = ((value ^ 0xFF + 1) & 0xFF) >> 2
				SNR = int(-value)
			} else {
				//divide by 4
				SNR = int((value & 0xFF) >> 2)
			}

			if sx1272 {
				rssicorr = 139
			} else {
				rssicorr = 157
			}

			fmt.Printf("Packet RSSI: %d, ", int(go_readReg(0x1A))-rssicorr)
			fmt.Printf("RSSI: %d, ", int(go_readReg(0x1B))-rssicorr)
			fmt.Printf("SNR: %v, ", SNR)
			fmt.Printf("Length: %v", int(receivedbytes))
			fmt.Printf("\n")
			fmt.Printf("Payload: %s\n", message)

		} //received a message
	} // dio0=1
}

func go_receive() (bool, string) {
	var payload [256]byte

	// clear rxDone
	go_writeReg(go_REG_IRQ_FLAGS, 0x40)

	irqflags := int(go_readReg(go_REG_IRQ_FLAGS))

	// payload crc: 0x20
	if (irqflags & 0x20) == 0x20 {
		fmt.Println("CRC error")
		go_writeReg(go_REG_IRQ_FLAGS, 0x20)
		return false, string(payload[:])
	} else {
		currentAddr := byte(go_readReg(go_REG_FIFO_RX_CURRENT_ADDR))
		receivedCount := byte(go_readReg(go_REG_RX_NB_BYTES))
		receivedbytes = receivedCount

		go_writeReg(go_REG_FIFO_ADDR_PTR, currentAddr)

		for i := 0; i < int(receivedCount); i++ {
			payload[i] = byte(go_readReg(go_REG_FIFO))
		}

	}
	return true, string(payload[:])
}

func main_func() {
        err:=rpio.Open()
        if err != nil {
          log.Fatal(err)
        }
        rpio.PinMode(ssPin,rpio.Output)
        rpio.PinMode(dio0,rpio.Input)
        rpio.PinMode(RST,rpio.Output)
        if err := rpio.SpiBegin(rpio.Spi0); err != nil {
		panic(err)
        }
        rpio.SpiChipSelect(CHANNEL) 
        rpio.SpiSpeed(500000)

	go_SetupLoRa()

	// Prepare to send block
	go_opmodeLora()
	go_opmode(go_OPMODE_STANDBY)
	go_writeReg(go_RegPaRamp, (go_readReg(go_RegPaRamp)&0xF0)|0x08) // set PA ramp-up time 50 uSec
	go_configPower(23)

	// Prepare to receive
	go_opmode(go_OPMODE_RX)
	// fmt.Printf("Listening at SF%d on %f Mhz.\n", go_sf, float64(float64(go_freq)/1000000))
	// Start MAIN CYCLE
	for {
		select {
		case <-send_signal:
			go_txlora(send_message)
			//fmt.Printf("Send packets at SF%d on %f Mhz.\n", go_sf, float64(float64(go_freq)/1000000))

			// return transciever to receive mode
			// set the IRQ mapping DIO0=TxDone DIO1=NOP DIO2=NOP
			go_writeReg(go_RegDioMapping1, go_MAP_DIO0_LORA_TXDONE|go_MAP_DIO1_LORA_NOP|go_MAP_DIO2_LORA_NOP)
			// clear all radio IRQ flags
			go_writeReg(go_REG_IRQ_FLAGS, 0xFF)
			// ari
			// mask all IRQs WITH TxDone
			go_writeReg(go_REG_IRQ_FLAGS_MASK, go_IRQ_LORA_TXDONE_MASK)
			go_opmode(go_OPMODE_RX)

		default:
			go_receivepacket()
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func main() {
	if len(os.Args[1:]) != 2 {
		fmt.Printf("Usage: %v <time to send> <string to send> \n", os.Args[0])
		os.Exit(0)
	}

	send_message = os.Args[2]
	time_from_arg, _ := strconv.Atoi(os.Args[1])
	send_signal = time.Tick(time.Duration(time_from_arg) * time.Second)

	main_func()

}
