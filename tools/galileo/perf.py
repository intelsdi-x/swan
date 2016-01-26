class Perf:
	def __init__(self, command):
		self.command = command

	def __str__(self):
		return "perf stat -o perf.txt " + str(self.command)
