package isolation

import "os/exec"
import "io/ioutil"
import "strconv"


type MemorySizeShares struct {
 cgroupName string
 memorySizeShares string

}

func NewMemorySizeShares(nameOfTheCgroup string, memorySizeShares string ) *MemorySizeShares{
	return &MemorySizeShares{cgroupName: nameOfTheCgroup, memorySizeShares: memorySizeShares}
}

func (memorySize *MemorySizeShares) Create() error {
	return nil
}

func (memorySize *MemorySizeShares) Delete() error {

	controllerName := "memory"
	cgroupName := "A"
        cmd := exec.Command("sh", "-c", "cgdelete -g "+controllerName+":"+cgroupName)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}


func (memorySize *MemorySizeShares) Isolate(PID int) error {

     	// 1.a Create memory size cgroup
	controllerName := "memory"
	cgroupName := "A"


        cmd := exec.Command("sh", "-c", "cgcreate -g memory:"+cgroupName)

	err := cmd.Run()
	if err != nil {
		return err
	}

     	// 1.b Set memory size cgroup shares

	memoryShares := "512M"

        cmd = exec.Command("sh", "-c", "cgset -r memory.limit_in_bytes="+memoryShares+" "+cgroupName)

	err = cmd.Run()
	if err != nil {
		return err
	}

     	// 4. Set PID to cgroups

	//Associate task with the cgroup
	//cgclassify seems to exit with error so temporarily using file io

	controllerName = "memory"

        strPID :=strconv.Itoa(PID)
	d := []byte(strPID)
	err3 := ioutil.WriteFile("/sys/fs/cgroup/"+controllerName+"/"+cgroupName+"/tasks", d, 0644)

	if err3 != nil {
		panic(err3.Error())
	}




//        cmd = exec.Command("sh", "-c", "cgclassify -g memory:A " + string(PID))
//
//	err = cmd.Run()
//	if err != nil {
//		panic("Cgclassify failed: " + err.Error())
//	}


	return nil
}


//	//Write CPU shares
//        strShares :=strconv.Itoa(1024)
//	c := []byte(strShares)
//	err2 := ioutil.WriteFile("/tmp/x.mri", c, 0644)
//
//	if err2 != nil {
//		panic(err2.Error())
//	}
//
//	//Associate task with the cgroup
//        strPID :=strconv.Itoa(PID)
//	d := []byte(strPID)
//	err3 := ioutil.WriteFile("/tmp/y.mri", d, 0644)
//
//	if err3 != nil {
//		panic(err3.Error())
//	}
//
//
//
//
//
//func (cpu *CpuShares) Isolate(PID int) error {
//        cmd := exec.Command("sh", "-c", "touch /tmp/x.tmp")
//
//	err := cmd.Run()
//	if err != nil {
//		panic(err.Error())
//	}
//
//	return nil
//}
