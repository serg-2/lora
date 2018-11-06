#!/bin/bash
wiring_pi_dir=/home/pi/wiringPi/wiringPi

gcc -Wall -c mainlib.c -o mainlib.o
ar -rcs libmainlib.a mainlib.o $wiring_pi_dir/wiringPi.o $wiring_pi_dir/wiringPiSPI.o $wiring_pi_dir/softPwm.o $wiring_pi_dir/softTone.o $wiring_pi_dir/piHiPri.o
go build -ldflags "-linkmode external -extldflags -static" lora.go

