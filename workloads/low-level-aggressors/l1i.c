#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#define ICPRESSURE do { x ^= y; } while (0)

#define ICPRESSURE2      ICPRESSURE;       ICPRESSURE
#define ICPRESSURE4      ICPRESSURE2;      ICPRESSURE2
#define ICPRESSURE8      ICPRESSURE4;      ICPRESSURE4
#define ICPRESSURE16     ICPRESSURE8;      ICPRESSURE8
#define ICPRESSURE32     ICPRESSURE16;     ICPRESSURE16
#define ICPRESSURE64     ICPRESSURE32;     ICPRESSURE32
#define ICPRESSURE128    ICPRESSURE64;     ICPRESSURE64
#define ICPRESSURE256    ICPRESSURE128;    ICPRESSURE128
#define ICPRESSURE512    ICPRESSURE256;    ICPRESSURE256
#define ICPRESSURE1024   ICPRESSURE512;    ICPRESSURE512
#define ICPRESSURE2048   ICPRESSURE1024;   ICPRESSURE1024
#define ICPRESSURE4096   ICPRESSURE2048;   ICPRESSURE2048
#define ICPRESSURE8192   ICPRESSURE4096;   ICPRESSURE4096
#define ICPRESSURE16384  ICPRESSURE8192;   ICPRESSURE8192
#define ICPRESSURE32768  ICPRESSURE16384;  ICPRESSURE16384
#define ICPRESSURE65536  ICPRESSURE32768;  ICPRESSURE32768
#define ICPRESSURE131072 ICPRESSURE65536;  ICPRESSURE65536
#define ICPRESSURE262144 ICPRESSURE131072; ICPRESSURE131072

int main(int argc, char **argv) {
  int x = 0xf0;
  int y = 0x0f;

  int usr_timer = atoi(argv[1]);
  printf ("%d\n", usr_timer); 

  clock_t begin, end;
  double time_spent;

  begin = clock();

  int intensity = 0; 

  intensity = atoi(argv[2]);

  printf("%d\n", intensity); 

  //for (int j = 0; j < 10*usr_timer; j++) { 
  for (int j = 0; j < 1500; j++) { 
	switch (intensity) {
	    case 0:  for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE; } //sleep(0.1); } //intensity = 1; 
	    case 1:  for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE2; } //sleep(0.1); } //intensity = 2;
	    case 2:  for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE4; } //intensity = 3; //while(1) { ICPRESSURE4; }
	    case 3:  for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE8; } //intensity = 4;
	    case 4:  for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE16; } //intensity = 5;
	    case 5:  for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE32; } //intensity = 6;
	    case 6:  for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE64; } //intensity = 7;
	    case 7:  for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE128; } //intensity = 8;
	    case 8:  for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE256; } //intensity = 9;
	    case 9:  for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE512; } //intensity = 10;
	    case 10: for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE1024; } //intensity = 11;
	    case 11: for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE2048; } //intensity = 12;
	    case 12: for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE4096; } //intensity = 13;
	    case 13: for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE8192; } //intensity = 14;
	    case 14: for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE16384; } //intensity = 15;
	    case 15: for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE32768; } //intensity = 16;
	    case 16: for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE65536; } //intensity = 17;
	    case 17: for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE131072; } //intensity = 18;
	    case 18: for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE262144; } //intensity = 19;
	    case 19: for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE262144; } //intensity = 20;
	    case 20: for (int i = 0; i < int(usr_timer); i++) { ICPRESSURE262144; } 
  	}
  }

  end = clock();
  time_spent = (double)(end - begin) / CLOCKS_PER_SEC;
  printf("Time spent: %f\n", time_spent);
  
  return 0;
}
