#include "mainlib.c"

int main (int argc, char *argv[]) {

    if (argc < 2) {
        printf ("Usage: argv[0] sender|rec [message]\n");
        exit(1);
    }

    wiringPiSetup () ;
    pinMode(ssPin, OUTPUT);
    pinMode(dio0, INPUT);
    pinMode(RST, OUTPUT);

    wiringPiSPISetup(CHANNEL, 500000);

    SetupLoRa();

    if (!strcmp("sender", argv[1])) {
        opmodeLora();
        // enter standby mode (required for FIFO loading))
        opmode(OPMODE_STANDBY);

        writeReg(RegPaRamp, (readReg(RegPaRamp) & 0xF0) | 0x08); // set PA ramp-up time 50 uSec

        configPower(23);

        printf("Send packets at SF%i on %.6lf Mhz.\n", sf,(double)freq/1000000);
        printf("------------------\n");

        char buf[100];        

	if (argc == 2)
	    strncpy((char *)buf, "test", sizeof("test"));
        if (argc > 2)
            strncpy((char *)buf, argv[2], sizeof(buf));

        while(1) {
            txlora((uint8_t*)buf, 20);
            delay(2000);
        }
    } else {

        // radio init
        opmodeLora();
        opmode(OPMODE_STANDBY);
        opmode(OPMODE_RX);
        printf("Listening at SF%i on %.6lf Mhz.\n", sf,(double)freq/1000000);
        printf("------------------\n");
        while(1) {
            receivepacket(); 
            delay(1);
        }

    }

    return (0);
}
