default: memCap memBench memCapAlt memBw l1i l1d l3 stream.100M

memCap:
	g++ -o memCap memCap.c

memBench:
	g++ -o memBench memBench.c

memCapAlt:
	g++ -o memCapAlt memCapAlt.c

memBw:
	g++ -o memBw memBw.c

l1i:
	g++ -o l1i l1i.c

l1d:
	g++ -o l1d l1d.c -lrt

l3:
	g++ -o l3 l3.c -lrt

stream.100M:
	gcc -O2 -fopenmp -DTUNED -DVERBOSE -DNTIMES=500 -DSTREAM_ARRAY_SIZE=100000000 stream.c -o stream.100M

clean:
	rm memCap
	rm memBench
	rm memCapAlt
	rm memBw
	rm l1i
	rm l1d
	rm l3
	rm stream.100M
