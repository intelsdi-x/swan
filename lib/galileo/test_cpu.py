import cpu

def main():
    cpus = cpu.Cpus()

    # Choose one socket
    sockets = cpus.unique_sockets(1)

    print "Got %d sockets" % len(sockets)

    socket = sockets[0]

    print "Selected socket %d" % socket.id

    cores = socket.unique_cores(3)

    print "Got %d cores:" % len(cores)

    print "Selected 3 cores: %d %d %d" % (cores[0].id, cores[1].id, cores[2].id)

    hyper_thread_ids = []
    for hyper_thread_id in cores[0].hyper_threads.keys():
        hyper_thread_ids.append(hyper_thread_id)

    print "Hyper threads on core %d: %d %d" % (cores[0].id, hyper_thread_ids[0], hyper_thread_ids[1])

if __name__ == "__main__":
    main()
