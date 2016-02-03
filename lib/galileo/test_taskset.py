import unittest
from taskset import Taskset


class ExperimentTest(unittest.TestCase):
    def test_empty_cpus(self):
        self.assertEqual(str(Taskset([], "foobar")), "foobar")

    def test_one_cpu(self):
        self.assertEqual(str(Taskset(["1"], "foobar")), "taskset -c 1 foobar")

    def test_multiple_cpus(self):
        self.assertEqual(str(Taskset(["1", "2"], "foobar")), "taskset -c 1,2 foobar")