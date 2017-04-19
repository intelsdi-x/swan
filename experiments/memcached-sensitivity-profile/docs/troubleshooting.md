# Troubleshooting

## Unstable Experiment results

In general, look at all running processes at the mutilate master, agents and on the target machine. Try to reduce the number of processes running at any time to reduce the likelihood of interference.
memcached and mutilate are sensitive to processes which use any network bandwidth and otherwise may interfere with normal execution speed. Example of these are tracing tools like `iftop`. Therefore, be cautious using instrumentation tools while conducting experiments.

## Known issues with mutilate

There may be issues with synchronization between the master and agents. If any of master or agents reports `out of sync`, all processes and the measurement have to be restarted.

If you see `connection closed by peer` on the agents, this is most likely related to the SYN cookies mentioned above.
