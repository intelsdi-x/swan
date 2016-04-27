package isolation

import "os/exec"
import "io/ioutil"
import "strconv"


type CpuSetShares struct {
 cgroupName string
 cpuSetShares string

}

func NewCpuSetShares(nameOfTheCgroup string, myCpuSetShares string ) *CpuSetShares{
	return &CpuSetShares{cgroupName: nameOfTheCgroup, cpuSetShares: myCpuSetShares}
}

func (cpuSet *CpuSetShares) Create() error {
	return nil
}

func (cpuSet *CpuSetShares) Delete() error {

	controllerName := "cpuset"
	cgroupName := "A"
        cmd := exec.Command("sh", "-c", "cgdelete -g "+controllerName+":"+cgroupName)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}


func (cpuSet *CpuSetShares) Isolate(PID int) error {
	cgroupName := "B"

     	// 2.a Create cpuset cgroup

	controllerName := "cpuset"


        cmd := exec.Command("sh", "-c", "cgcreate -g "+controllerName+":"+cgroupName)

	err := cmd.Run()
	if err != nil {
		return err
	}

     	// 2.b Set cpuset cgroup cpus

	cgCpus := "0-3"

        cmd = exec.Command("sh", "-c", "cgset -r cpuset.cpus="+cgCpus+" "+cgroupName)

	err = cmd.Run()
	if err != nil {
		return err
	}

	
	cmd = exec.Command("sh", "-c", "cgset -r cpuset.mems=0-1 "+cgroupName)

        err = cmd.Run()
        if err != nil {
                return err
        }

     	// 3. Set PID to cgroups

	//Associate task with the cgroup
	//cgclassify seems to exit with error so temporarily using file io

	controllerName = "cpuset"

        strPID :=strconv.Itoa(PID)
	d := []byte(strPID)
	err3 := ioutil.WriteFile("/sys/fs/cgroup/"+controllerName+"/"+cgroupName+"/tasks", d, 0644)

	if err3 != nil {
		panic(err3.Error())
	}




//        cmd = exec.Command("sh", "-c", "cgclassify -g cpuset:A " + string(PID))
//
//	err = cmd.Run()
//	if err != nil {
//		panic("Cgclassify failed: " + err.Error())
//	}


	return nil
}
