default: 
	g++ -o memCap memCap.c
	g++ -o memBench memBench.c
	g++ -o memCapAlt memCapAlt.c
	g++ -o memBw memBw.c
	g++ -o l1i l1i.c -lrt
	g++ -o l1d l1d.c -lrt
	g++ -o l3 l3.c -lrt
	gcc -O2 -fopenmp -DNTIMES=500 -DSTREAM_ARRAY_SIZE=100000000 stream.c -o stream.100M

clean: 
	rm memCap
	rm memBench
	rm memCapAlt
	rm memBw
	rm l1i
	rm l1d
	rm l3
	rm stream.100M
