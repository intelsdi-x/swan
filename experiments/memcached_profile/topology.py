from cpu import Cpus

# TODO(nnielsen): Parameterize number of victim and workload hyper threads.
def generate_topology(aggressor=False, aggressor_on_hyper_threads=True, aggressor_on_core=True, aggressor_on_socket=True):
    cpus = Cpus()

    #
    # Determine core affinity
    #
    workload_hyper_thread_id = 0
    workload_memory_node_id = 0
    victim_hyper_thread_id = 0
    victim_memory_node_id = 0
    aggressor_hyper_thread_id = 0
    aggressor_memory_node_id = 0
    if len(cpus.sockets) < 2:
        log.warning("Single socket experiment not yet supported")
        sys.exit(1)

    # Try to separate on different sockets.
    sockets = cpus.unique_sockets(2)
    workload_socket = sockets[0]
    victim_socket = sockets[1]

    # Gross assumption: memory node == socket id. Validate with hwloc.
    # We don't know which memory node to attach aggressor to yet.
    workload_memory_node_id = workload_socket.id
    victim_memory_node_id = victim_socket.id 

    workload_cores = workload_socket.unique_cores(2)
    workload_hyper_thread_id = workload_cores[0].unique_hyper_threads(1)[0].id

    victim_cores = victim_socket.unique_cores(2)
    victim_core = victim_cores[0]
    victim_hyper_threads = victim_core.unique_hyper_threads(2)

    # Parent cgroup need to cover all used hyper threads. Make as 'set' to avoid duplicates. 
    parent_hyper_threads = set([str(workload_hyper_thread_id), str(victim_hyper_thread_id)])
    parent_mems = set([str(workload_memory_node_id), str(victim_memory_node_id)])

    # Aggressor configuration is hoisted to make them available in the cgroups setup below.
    aggressor_memory_node_id = 0
    aggressor_hyper_thread_id = victim_hyper_threads[0].id

    if aggressor == True:
        # TODO(nnielsen): Flatten the logic below.
        if aggressor_on_hyper_threads == True:
            # If aggressor should be placed on same hyper thread, reuse hyper thread id.
            aggressor_hyper_thread_id = victim_hyper_threads[0].id
            aggressor_memory_node_id = victim_memory_node_id
        else:
            if aggressor_on_core == True:
                # If aggressor should not be placed on same hyper thread, but same core. Use
                # alternative hyper thread id generated above.
                aggressor_hyper_thread_id = victim_hyper_threads[1].id
                aggressor_memory_node_id = victim_memory_node_id
            else:
                # If aggressor should not be placed on same core, but on same socket. Use
                # alternative core generated above.
                if aggressor_on_socket == True:
                    aggressor_hyper_thread_id = victim_cores[1].unique_hyper_threads(1)[0].id
                    aggressor_memory_node_id = victim_memory_node_id
                else:
                    # If aggressor should not be placed on same socket, place on hyper thread on a
                    # different core than the workload.
                    aggressor_hyper_thread_id = workload_cores[1].unique_hyper_threads(1)[0].id
                    aggressor_memory_node_id = workload_memory_node_id

        parent_hyper_threads.add(str(aggressor_hyper_thread_id))
        parent_mems.add(str(aggressor_memory_node_id))

    # Parent cgroup need to cover all hyper threads. We therefore need to figure out which aggressor
    # hyper threads that will be used before listing the cgroups settings.
    output = [
        "/memcached_experiment/cpuset.cpus=%s" % (",".join(parent_hyper_threads)),
        "/memcached_experiment/cpuset.mems=%s" % (",".join(parent_mems)),
        "/memcached_experiment/workload/cpuset.cpus=%s" % workload_hyper_thread_id,
        "/memcached_experiment/workload/cpuset.mems=%s" % workload_memory_node_id,
        "/memcached_experiment/victim/cpuset.cpus=%s" % victim_hyper_thread_id,
        "/memcached_experiment/victim/cpuset.mems=%s" % victim_memory_node_id,
    ]

    # To preserve order in creation of the cgroups, add aggressor cgroups if enabled.
    if aggressor == True:
        output += [
            "/memcached_experiment/aggressor/cpuset.cpus=%s" % aggressor_hyper_thread_id,
            "/memcached_experiment/aggressor/cpuset.mems=%s" % aggressor_memory_node_id
        ]

    return output
