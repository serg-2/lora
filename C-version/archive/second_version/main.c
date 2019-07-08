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

        if (argc > 2)
            strncpy((char *)hello, argv[2], sizeof(hello));

//ari3
        int fd = open("test.txt", O_NONBLOCK);
	if (fd == -1) {
	     printf("Unable to open file\n");
	     return 1;
	}
//ari4
	 int flags = fcntl(fd, F_GETFL);
	 if (flags & O_NONBLOCK) {
	     printf("non block is set\n");
	}
        
        while(1) {
//ari5
            char buf[100];
            lseek (fd, -4,SEEK_END);
            read (fd,buf,4);
            txlora((uint8_t*)buf,4);
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
