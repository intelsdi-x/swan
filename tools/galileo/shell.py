import glog as log
import subprocess
import time

class Shell:
	def __init__(self, commands, output="output"):
		processes = {}

		for command in commands:
			output_file = open(output, "a+")
			log.info("started command: '" + str(command) + "'")
			p = subprocess.Popen(["sh", "-c", str(command)], stdout=output_file, stderr=output_file)
			processes[p.pid] = {'process': p, 'output_file': output_file}
		
		running = processes
		while len(running) is not 0:
			exited_pids = []
			for pid, process_obj in running.iteritems():
				process = process_obj['process']
				status = process.poll()
				if status is not None:
					log.info("ended command: '" + str(command) + "' with status code " + str(status))
					outpuf_file = process_obj['output_file']
					output_file.flush()
					exited_pids.append(pid)

					# TODO: Write post mortem to log

			for pid in exited_pids:
				del running[pid]

			time.sleep(0.1)
