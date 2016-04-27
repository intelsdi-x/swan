# Swan vision and motivation

Swan is an experiment and evaluation methodology for optimizing cloud schedulers.
Through coordination of distributed testing, Swan provides a framework to capture real cloud workloads and ability for developers and operators to gain deep insight into their workload behavior in a controlled environment.
Swan emphasizes the need for _co-located_ workloads experiments.
Modern infrastructures run tens to hundreds of tasks per server and the different combinations of how those get placed on a server turns out to be a huge problem space to uncover.
We treat this exploration as a workload in itself.

## Principles

Experiment setups can be complex and involve standing up and configuring services, software and hardware, and span many machines.
For that reason, and because of the parallel opportunity, experiments are _written in code_ and swan provides library for running distributed experiments under tightly controlled conditions.
The experiment abstractions does not try to shoehorn a particular format for the output or analysis, but provides just enough help experiment developers make this happen.
