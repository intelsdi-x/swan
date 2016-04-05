package runner

type Task interface{
	stop()
	status()
	output()
}
