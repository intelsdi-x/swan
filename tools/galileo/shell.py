import glog as log
import subprocess

class Shell:
	def __init__(self, command, output="output"):
		output_file = open(output, "a+")
		log.info("started command: '" + str(command) + "'")
		p = subprocess.Popen(["sh", "-c", str(command)], stdout=output_file, stderr=output_file)
		p.wait()
		log.info("ended command: '" + str(command) + "'")
		output_file.flush()
		
