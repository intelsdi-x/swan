all: deps build run

run:
	(cd experiments/memcached_profile/; sudo -E python main.py )

deps:
	pip install -r requirements.txt

build: build-workloads

build-workloads: 
	(cd workloads/data_caching/memcached; ./build.sh)
	(cd workloads/low-level-aggressors/; make)

