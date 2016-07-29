#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#define ICPRESSURE1 do { x ^= y; } while (0)

#define ICPRESSURE2      ICPRESSURE1;      ICPRESSURE1
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
#define ICPRESSURE524288 ICPRESSURE262144; ICPRESSURE262144

#define INTENSITY(X) ICPRESSURE ## X
#define REPEAT(N) for (int iteration = 0; (N == infinite) || (iteration < N); iteration++)

int main(int argc, char **argv) {
  if ((argc < 2) || (argc > 3)) {
    fprintf(stderr, "usage: l1i <intensity 1-20(highest)> (<iterations>)\n");
    return 1;
  }

  clock_t begin = clock();

  const int infinite = -1;
  int iterations = infinite;

  if (argc == 3) {
    iterations = atoi(argv[2]);
    printf("iterations: %d\n", iterations);
  } else {
    printf("iterations: infinite\n");
  }

  int intensity = 0;
  intensity = atoi(argv[1]);
  printf("intensity: %d\n", intensity);

  int x = 0xf0;
  int y = 0x0f;

	switch (intensity) {
      case 0:  { REPEAT(iterations) { INTENSITY(1);      } } break;
      case 1:  { REPEAT(iterations) { INTENSITY(2);      } } break;
      case 2:  { REPEAT(iterations) { INTENSITY(4);      } } break;
      case 3:  { REPEAT(iterations) { INTENSITY(8);      } } break;
      case 4:  { REPEAT(iterations) { INTENSITY(16);     } } break;
      case 5:  { REPEAT(iterations) { INTENSITY(32);     } } break;
      case 6:  { REPEAT(iterations) { INTENSITY(64);     } } break;
      case 7:  { REPEAT(iterations) { INTENSITY(128);    } } break;
      case 8:  { REPEAT(iterations) { INTENSITY(256);    } } break;
      case 9:  { REPEAT(iterations) { INTENSITY(512);    } } break;
      case 10: { REPEAT(iterations) { INTENSITY(1024);   } } break;
      case 11: { REPEAT(iterations) { INTENSITY(2048);   } } break;
      case 12: { REPEAT(iterations) { INTENSITY(4096);   } } break;
      case 13: { REPEAT(iterations) { INTENSITY(8192);   } } break;
      case 14: { REPEAT(iterations) { INTENSITY(16384);  } } break;
      case 15: { REPEAT(iterations) { INTENSITY(32768);  } } break;
      case 16: { REPEAT(iterations) { INTENSITY(65536);  } } break;
      case 17: { REPEAT(iterations) { INTENSITY(131072); } } break;
      case 18: { REPEAT(iterations) { INTENSITY(131072); } } break;
      case 19: { REPEAT(iterations) { INTENSITY(262144); } } break;
      case 20: { REPEAT(iterations) { INTENSITY(524288); } } break;
  }

  clock_t end = clock();
  double time_spent = (double)(end - begin) / CLOCKS_PER_SEC;
  printf("time spent: %f seconds\n", time_spent);

  return 0;
}
