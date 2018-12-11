#!/bin/bash

wiring_pi_dir=/home/pi/wiringPi/wiringPi

ar -rcs libmainlib.a $wiring_pi_dir/wiringPi.o $wiring_pi_dir/wiringPiSPI.o $wiring_pi_dir/softPwm.o $wiring_pi_dir/softTone.o $wiring_pi_dir/piHiPri.o

go build -ldflags "-linkmode external -extldflags -static" lora.go

