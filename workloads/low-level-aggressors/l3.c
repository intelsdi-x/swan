#include <sys/mman.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <time.h>
#include <sched.h>

// Returns the size of the L3 cache in bytes, otherwise -1 on error.
int cache_size() {
    // We grab the cache size from cpu 0, assuming that all cpu caches are the
    // same size.
    const char* cache_size_path =
        "/sys/devices/system/cpu/cpu0/cache/index3/size";

	FILE *cache_size_fd;
	if (!(cache_size_fd = fopen(cache_size_path, "r"))) {
		perror("could not open cache size file");
		return -1;
	}

	char line[512];
	if(!fgets(line, 512, cache_size_fd)) {
		fclose(cache_size_fd);
		perror("could not read from cache size file");
		return -1;
	}

	// Strip newline
	const int newline_pos = strlen(line) - 1;
	if (line[newline_pos] == '\n') {
		line[newline_pos] = '\0';
	}

	// Get multiplier
	int multiplier = 1;
	const int multiplier_pos = newline_pos - 1;
	switch (line[multiplier_pos]) {
		case 'K':
			multiplier = 1024;
		break;
		case 'M':
			multiplier = 1024 * 1024;
		break;
		case 'G':
			multiplier = 1024 * 1024 * 1024;
		break;
	}

	// Remove multiplier
	if (multiplier != 1) {
		line[multiplier_pos] = '\0';
	}

	// Line should now be clear of non-numeric characters
	int value = atoi(line);

	int cache_size = value * multiplier;

	fclose(cache_size_fd);

	return cache_size;
}

int main(int argc, char **argv) {
	char* volatile block;
	int CACHE_SIZE = cache_size(); 
	printf("Detected L3 cache size: %d bytes\n", CACHE_SIZE);

	/*Usage: ./l3 <duration in sec>*/
	if (argc < 2) { 
		printf("Usage: ./l3 <duration in sec>\n"); 
		exit(0); 
	}	
	block = (char*)mmap(NULL, CACHE_SIZE, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, 0, 0);

	int usr_timer = atoi(argv[1]);
	double time_spent = 0.0; 
  	clock_t begin, end;


	while (time_spent < usr_timer) {
  		begin = clock();
		memcpy(block, block+CACHE_SIZE/2, CACHE_SIZE/2);
		// note: replaced original throttling sleep with yielding that gives chance ther workloads to run
		sched_yield(); // sleep((float)(usr_timer-time_spent)/usr_timer);
		end = clock();
  		time_spent += (double)(end - begin) / CLOCKS_PER_SEC;
	}
	return 0;
}

