package isolation

import "os/exec"
import "io/ioutil"
import "strconv"

//MemorySize defines input data
type MemorySize struct {
 cgroupName string
 memorySize string

}

//NewMemorySizeShares creates an instance of input data
func NewMemorySize(nameOfTheCgroup string, memorySizeShares string ) *MemorySize{
	return &MemorySize{cgroupName: nameOfTheCgroup, memorySize: memorySizeShares}
}

//Delete removes specified cgroup
func (memorysize *MemorySize) Delete() error {

        cmd := exec.Command("sh", "-c", "cgdelete -g memory:"+memorysize.cgroupName)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

//Isolate create specified cgroup and associates specified process id
func (memorysize *MemorySize) Isolate(PID int) error {

     	// 1.a Create memory size cgroup

        cmd := exec.Command("sh", "-c", "cgcreate -g memory:"+memorysize.cgroupName)

	err := cmd.Run()
	if err != nil {
		return err
	}

     	// 1.b Set cgroup memory size 

        cmd = exec.Command("sh", "-c", "cgset -r memory.limit_in_bytes="+memorysize.memorySize+" "+memorysize.cgroupName)

	err = cmd.Run()
	if err != nil {
		return err
	}

     	// 2. Set PID to cgroups

	//Associate task with the cgroup
	//cgclassify and cgexec seem to exit with error so temporarily using file io


        strPID :=strconv.Itoa(PID)
	d := []byte(strPID)
	err = ioutil.WriteFile("/sys/fs/cgroup/memory/"+memorysize.cgroupName+"/tasks", d, 0644)

	if err != nil {
		panic(err.Error())
	}




//      cmd = exec.Command("sh", "-c", "cgclassify -g memory:A " + string(PID))
//
//	err = cmd.Run()
//	if err != nil {
//		panic("Cgclassify failed: " + err.Error())
//	}


	return nil
}
