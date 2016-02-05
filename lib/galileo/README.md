# Galileo

Galileo works as a driver for co-located workload experiments.
The intent is to focus on systematic testing and produce self-contained structured output  as an intermediate representation for data processing.

The end goal is to be able to produce:
 - Sensitivity Profiles for latency sensitive workloads.
 - Exercise real colocations and generate oversubscription quality scores (OQS) for oversubscription policies.

## Installation

While no `setup.py` is shipped with galileo yet, you can install the dependencies by running:

```
$ sudo yum install python-devel freetype-devel libpng-devel
$ sudo easy_install pip 
$ sudo pip install -r requirements.txt
```

## Usage

While galileo is a library to be imported and extended by the user, you can run the bundled example experiments:

```
$ python example_simple.py
$ python example_cgroup.py
```

## Development

For API documentation, query with `pydoc <module>`. For example:
```
$ pydoc cgroup
Help on module cgroup:

NAME
    cgroup

FILE
    /Users/nqnielse/workspace/scheduler-workloads/lib/galileo/cgroup.py

CLASSES
    Cgroup
    
    class Cgroup
     |  Cgroup hierarchy control class: creates and configures cgroups hierarchies based on a set of 'desired states'.
     |  See __init__() documentation for format of desired states.
     |  
     |  Methods defined here:
     |  
     |  __init__(self, desired_states, create_func=None, destroy_func=None, set_func=None)
     |      :param desired_states: List of desired states in form of <path>=<value>.
     |                             For example, ["/A/B/cpu.shares=1", "/A/C/mem.limit_in_bytes=512"]
     |                             The above will create:
     |                             1) The 'A' root of the two children with 'cpu' and 'mem' cgroups
     |                             2) 'B' child cgroup with 'cpu' and 'mem' and set cpu shares to 1
     |                             3) 'C' child cgroup with 'cpu' and 'mem' and set memory limit to 512 bytes
     |      
     |                             The settings will be applied in order, so settings in the parents can be set before the
     |                             children's settings are applied. This is necessary for setting 'cpuset's.
     |      :param create_func:    Function to use for creating cgroups. For testing purposes only.
     |      :param destroy_func:   Function to use for destroy cgroups. For testing purposes only.
     |      :param set_func:       Function to use for setting cgroup parameters. For testing purposes only.
     |  
     |  destroy(self)
     |      Tears down cgroups created in __init__()
     |  
     |  execute(self, location, command)
     |      Executes command attached to cgroup defined in 'location'
     |      
     |      :param location:    Cgroup hierarchy to run command under.
     |                          For example '/A/B' for '/sys/fs/cgroup/[cpuset,cpu]/A/B'
     |      :param command:     Command to run
     |      :return:            String to execute with 'Shell()'
```

Run unit tests with:
```
py.test
```
