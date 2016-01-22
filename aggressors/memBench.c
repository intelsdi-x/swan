#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include <signal.h>

#include <time.h>
#include <sys/time.h>
#include <stdio.h>

//#define ARR_BYTES (1<<12) //4KB
//#define ARR_BYTES (1<<14) //16KB
//#define ARR_BYTES (1<<17) //128KB
//#define ARR_BYTES (1<<21) //2MB
//#define ARR_BYTES (1<<26) //64MB

//#define ARR_BYTES (1<<29) //512MB
#define ARR_BYTES (1<<30) //1GB

char bytes[ARR_BYTES] __attribute__((aligned(64)));

void bypassingGccsOptimizer(uint64_t* x0, uint64_t* x1, uint64_t* x2, uint64_t* x3, uint64_t* x4, uint64_t* x5, uint64_t* x6, uint64_t* x7) __attribute__((noinline));

void bypassingGccsOptimizer(uint64_t* x0, uint64_t* x1, uint64_t* x2, uint64_t* x3, uint64_t* x4, uint64_t* x5, uint64_t* x6, uint64_t* x7) {
    *x0 = (uint64_t) bytes;
    *x1 = (uint64_t) (bytes+64);
    *x2 = (uint64_t) (bytes+128);
    *x3 = (uint64_t) (bytes+192);
    *x4 = (uint64_t) (bytes+256);
    *x5 = (uint64_t) (bytes+320);
    *x6 = (uint64_t) (bytes+384);
    *x7 = (uint64_t) (bytes+448);
}

int main( int argc, char **argv )
{
    double t1, t2, tzero;
    struct timeval tv;
    long int i = 0;

    long int loops = 1000;//0;
    double cpuspeed = 2301.000; //replace with the frequency of the tested CPU

    tzero = 0;

    printf("cpuspeed = %f\n", cpuspeed);
    printf("loops = %ld\n", loops);

    //Warm up
    uint32_t* vals = (uint32_t*) bytes;
    uint64_t nVals = (ARR_BYTES>>2);
    for (i = 0; i < nVals; i++) {
        vals[i] = i;
    }

    printf("warmed up\n");

    gettimeofday(&tv, NULL);
    t1 = (double)tv.tv_sec + (double)tv.tv_usec/1000000.0;

    double time_spent = 0; 
    int usr_timer = atoi(argv[1]);
    clock_t begin, end; 

    i = 0;
    while (time_spent < usr_timer) {
        i++;
#if 1 
        char* p0, *p1, *p2, *p3, *p4, *p5, *p6, *p7; 
        p0 = bytes;
        p1 = bytes+64;
        p2 = bytes+128;
        p3 = bytes+192;
        p4 = bytes+256;
        p5 = bytes+320;
        p6 = bytes+384;
        p7 = bytes+448;

        uint64_t j;
	begin = clock(); 
        for (j = 0; j < 64*ARR_BYTES; j+= 8*64) {
            //char* ptr = &bytes[j];
#define OPCODE "mov"
            //In GCC --> opcode  SRC, DST 
            asm volatile(
                OPCODE " (%0), %%r8 \n" //load
                OPCODE " (%1), %%r8 \n"
                OPCODE " (%2), %%r8 \n"
                OPCODE " (%3), %%r8 \n"
                OPCODE " (%4), %%r8 \n"
                OPCODE " (%5), %%r8 \n"
                OPCODE " (%6), %%r8 \n" //...
                OPCODE " (%7), %%r8 \n" //load
                :
                : "r" (p0), "r"(p1), "r" (p2), "r"(p3), "r" (p4), "r"(p5), "r" (p6), "r"(p7)
                : "r8" 
            );
            p0 += 64; p1 += 64; p2 += 64; p3 += 64;
            p4 += 64; p5 += 64; p6 += 64; p7 += 64;
        }
	end = clock();
  	time_spent += (double)(end - begin) / CLOCKS_PER_SEC;
    }
#endif

#if 0
#endif
    gettimeofday(&tv, NULL);
    t2 = (double)tv.tv_sec + (double)tv.tv_usec/1000000.0;
    uint64_t ldsPerLoop = (ARR_BYTES)/64;
    printf("%-40s = %#.5g \n", "reciprocal throughput" ,
            (cpuspeed * 1000000 / (double)i) * (t2 - t1 - tzero) / (double)ldsPerLoop);

    return 0;
}

