package ExperimentDriver

type Task interface {
	Wait() error    // Waits till Task finishes
	Status() int    //TBD: is running, not launched, stopped, finished?
	GetOutput() int //TBD - do I need this?
	Stop() error    //Forcefully stops the workload - waits till task stops
}

type Launcher interface {
	/**
	 * Launch Workload. The rest is hidden from Experiment Driver
	 * Returns:
	 * On error - (nil, error)
	 * On success - (Task, nil)
	 */
	Launch() (Task, error)
}

type LoadGenerator interface {
	/**
	 * Performs tunning.
	 * Takes 'slo' and timeout value as input.
	 * Returns (targetQPS,nil) on success (nil, error) otherwise
	 *
	 */
	Tune(slo int, timeout int) (targetQPS int, e error)
	//Launch workload generator with given 'rps' (Request per Second)
	// and for a given 'duration' of seconds (TBD).
	// Returns (sli, nil) on success (nil, err) otherwise.
	Load(rps int, duration int) (sli int, e error)
}

type ExperimentConfiguration struct {
	slo               int
	tuning_timeout    int
	load_duration     int
	load_points_count int
}

type SensitivitiProfileExperiment struct {
	lc Launcher
	lg LoadGenerator //for LC Task
	be []Launcher

	//
	tuning_timeout    int
	load_duration     int
	load_points_count int
	target_qps        int
	slo               int
	sli               [][]int
}

func (SensitivitiProfileExperiment) String() string {
	return "SensitivitiProfileExperiment type."
}

// Construct new SensitivityProfileExperiment object.
func NewSensitivitiProfileExperiment(c ExperimentConfiguration,
	pr_task Launcher, lg LoadGenerator, be []Launcher) *SensitivitiProfileExperiment {
	exp := SensitivitiProfileExperiment{}

	exp.slo = c.slo
	exp.tuning_timeout = c.tuning_timeout
	exp.load_duration = c.load_duration
	exp.load_points_count = c.load_points_count
	exp.lc = pr_task
	exp.lg = lg
	exp.be = be
	return &exp
}

// [internal]
// Runs single measurement of PR workload with given aggressor.
// Takes aggressor_no index in BE workload and specific loadPoint
// Return (sli, nil) on success (0, error) otherwise.
func (e *SensitivitiProfileExperiment) runMeasurement(aggressor_no int, loadPoint int) (sli int, err error) {

	lc_task, err := e.lc.Launch()
	if err != nil {
		return 0, err
	}

	//Run aggressor
	agr_task, err := e.be[aggressor_no].Launch()
	if err != nil {
		lc_task.Stop()
		return 0, err
	}
	//Run workload generator - blocking?
	//TBD: output? Raw data?
	sli, err = e.lg.Load(loadPoint, e.load_duration)

	lc_task.Stop()
	agr_task.Stop()
	return sli, err
}

// [internal]
// Executes single phase
//
func (e *SensitivitiProfileExperiment) runPhase(aggressor_no int) error {
	//Do we capture output from measurement?
	var err error

	//TBD - iteration
	for i := 5; i < 95; i += 5 {
		tmp_sli, err := e.runMeasurement(aggressor_no, i)
		if err != nil {
			return err
		}
		//Update sli only if measurement was successful
		e.sli[aggressor_no][i] = tmp_sli
	}
	return err
}

//Performs LC tunning
func (e *SensitivitiProfileExperiment) runTunning() error {
	pr_task, err := e.lc.Launch()
	if err != nil {
		return err
	}
	tmp_target_qps, err := e.lg.Tune(e.slo, e.tuning_timeout)
	if err == nil {
		e.target_qps = tmp_target_qps
	}
	pr_task.Stop()
	return err
}

// Prints nice output. TBD
func (e *SensitivitiProfileExperiment) PrintSensitivitiProfile() error {
	// TBD
	return nil
}

// Performs actual experiments
func (e *SensitivitiProfileExperiment) Run() error {

	var err error
	err = e.runTunning()
	if err != nil {
		//Stop here
		return err
	}

	for i, _ := range e.be {
		err = e.runPhase(i)
		if err != nil {
			return err
		}
	}
	//TBD
	e.PrintSensitivitiProfile()
	return nil
}
