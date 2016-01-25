import ga
import glog as log
import random

class SimpleExperiment(ga.Experiment):
	def __init__(self):
		ga.Experiment.__init__(self)

		def baseline(configuration):
			# Do work.

			return {
				"99.9 latency": 1.0 + random.uniform(0.0, 2.0)
			}

		def experiment(configuration):
			# Cassandra (2.0) + Kafka (3.0)
			return {
				"99.9 latency": 2.0 + random.uniform(0.0, 4.0)
			}

		self.add_phase("Baseline", baseline)
		self.add_phase("Experiment", experiment, [range(0, 8), range(100, 1000, 100)])

def main():
	s = SimpleExperiment()
	s.run()

if __name__ == "__main__":
	main()
