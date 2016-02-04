import unittest
from cpu import Cpus
import os

class CpuTest(unittest.TestCase):
    def test_single_cpu(self):
        cpus = Cpus("cpuinfo_examples/intel_core_i7_5557u_single.txt")
        self.assertEqual(len(cpus.hyper_threads), 1)
        self.assertEqual(len(cpus.sockets), 1)
        self.assertEqual(len(cpus.sockets[0].cores), 1)
        self.assertEqual(len(cpus.sockets[0].cores[0].hyper_threads), 1)

    def test_dual_cpu(self):
        cpus = Cpus("cpuinfo_examples/intel_core_i7_5557u_dual.txt")
        self.assertEqual(len(cpus.hyper_threads), 2)
        self.assertEqual(len(cpus.sockets), 1)
        self.assertEqual(len(cpus.sockets[0].cores), 2)
        self.assertEqual(len(cpus.sockets[0].cores[0].hyper_threads), 1)
        self.assertEqual(len(cpus.sockets[0].cores[1].hyper_threads), 1)

    def test_multi_socket_hyper_threaded_cpu(self):
        cpus = Cpus("cpuinfo_examples/intel_xeon_e5_2690v2_ht.txt")

        # Verify counts
        self.assertEqual(len(cpus.hyper_threads), 40)
        self.assertEqual(len(cpus.sockets), 2)
        for socket in cpus.sockets.values():
            self.assertEqual(len(socket.cores), 10)
            for core in socket.cores.values():
                self.assertEqual(len(core.hyper_threads), 2)

        # Verify topology
        self.assertEqual(cpus.sockets[0].cores[0].hyper_threads[0].id, 0)
        self.assertEqual(cpus.sockets[0].cores[1].hyper_threads[1].id, 1)
        self.assertEqual(cpus.sockets[0].cores[2].hyper_threads[2].id, 2)
        self.assertEqual(cpus.sockets[0].cores[3].hyper_threads[3].id, 3)
        self.assertEqual(cpus.sockets[0].cores[4].hyper_threads[4].id, 4)
        self.assertEqual(cpus.sockets[0].cores[8].hyper_threads[5].id, 5)
        self.assertEqual(cpus.sockets[0].cores[9].hyper_threads[6].id, 6)
        self.assertEqual(cpus.sockets[0].cores[10].hyper_threads[7].id, 7)
        self.assertEqual(cpus.sockets[0].cores[11].hyper_threads[8].id, 8)
        self.assertEqual(cpus.sockets[0].cores[12].hyper_threads[9].id, 9)
        self.assertEqual(cpus.sockets[1].cores[0].hyper_threads[10].id, 10)
        self.assertEqual(cpus.sockets[1].cores[1].hyper_threads[11].id, 11)
        self.assertEqual(cpus.sockets[1].cores[2].hyper_threads[12].id, 12)
        self.assertEqual(cpus.sockets[1].cores[3].hyper_threads[13].id, 13)
        self.assertEqual(cpus.sockets[1].cores[4].hyper_threads[14].id, 14)
        self.assertEqual(cpus.sockets[1].cores[8].hyper_threads[15].id, 15)
        self.assertEqual(cpus.sockets[1].cores[9].hyper_threads[16].id, 16)
        self.assertEqual(cpus.sockets[1].cores[10].hyper_threads[17].id, 17)
        self.assertEqual(cpus.sockets[1].cores[11].hyper_threads[18].id, 18)
        self.assertEqual(cpus.sockets[1].cores[12].hyper_threads[19].id, 19)
        self.assertEqual(cpus.sockets[0].cores[0].hyper_threads[20].id, 20)
        self.assertEqual(cpus.sockets[0].cores[1].hyper_threads[21].id, 21)
        self.assertEqual(cpus.sockets[0].cores[2].hyper_threads[22].id, 22)
        self.assertEqual(cpus.sockets[0].cores[3].hyper_threads[23].id, 23)
        self.assertEqual(cpus.sockets[0].cores[4].hyper_threads[24].id, 24)
        self.assertEqual(cpus.sockets[0].cores[8].hyper_threads[25].id, 25)
        self.assertEqual(cpus.sockets[0].cores[9].hyper_threads[26].id, 26)
        self.assertEqual(cpus.sockets[0].cores[10].hyper_threads[27].id, 27)
        self.assertEqual(cpus.sockets[0].cores[11].hyper_threads[28].id, 28)
        self.assertEqual(cpus.sockets[0].cores[12].hyper_threads[29].id, 29)
        self.assertEqual(cpus.sockets[1].cores[0].hyper_threads[30].id, 30)
        self.assertEqual(cpus.sockets[1].cores[1].hyper_threads[31].id, 31)
        self.assertEqual(cpus.sockets[1].cores[2].hyper_threads[32].id, 32)
        self.assertEqual(cpus.sockets[1].cores[3].hyper_threads[33].id, 33)
        self.assertEqual(cpus.sockets[1].cores[4].hyper_threads[34].id, 34)
        self.assertEqual(cpus.sockets[1].cores[8].hyper_threads[35].id, 35)
        self.assertEqual(cpus.sockets[1].cores[9].hyper_threads[36].id, 36)
        self.assertEqual(cpus.sockets[1].cores[10].hyper_threads[37].id, 37)
        self.assertEqual(cpus.sockets[1].cores[11].hyper_threads[38].id, 38)
        self.assertEqual(cpus.sockets[1].cores[12].hyper_threads[39].id, 39)
