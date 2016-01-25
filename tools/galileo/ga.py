import glog as log
import uuid
import numpy as np
import collections

class Experiment:
	def __init__(self):
		self.phases = collections.OrderedDict()

	def add_phase(self, name, phase, matrix=None):
		self.phases[name] = phase

	def run(self):
		# Create sandbox
		# Write system details (if available).

		run_id = str(uuid.uuid4())
		log.info("started experiment run '" + run_id + "'")

		for name, phase in self.phases.iteritems():
			results = []
			log.info("started phase '" + name + "'")
			for iteration in range(0, 3):
				log.info("started phase '" + name + "' iteration " + str(iteration))
				result = phase(None)
				results.append(result)
				log.info("  ended phase '" + name + "' iteration " + str(iteration))
			metrics = {}
			log.info("  ended phase '" + name + "'")
			for result in results:
				for metric_name, metric in result.iteritems():
					if metric_name not in metrics:
						metrics[metric_name] = []

					metrics[metric_name].append(metric)

			for metric_name, metrics in metrics.iteritems():
				log.info('phase \'' + name + '\' result: ' + metric_name + '(mean of ' + str(len(metrics)) + ' runs): ' + str(np.mean(metrics)))
				

		log.info("ended experiment run '" + run_id + "'")
