import unittest
from shell import Shell
import os


class ShellTest(unittest.TestCase):
    def test_empty_command(self):
        s = Shell("")
        self.assertEqual(len(s.processes), 0)

    def test_ls(self):
        s = Shell(["ls"], "/dev/null")
        self.assertEqual(len(s.processes), 1)
        self.assertEqual(s.processes.itervalues().next()['command'], 'ls')
        self.assertEqual(s.processes.itervalues().next()['status'], 0)

    def test_exit1(self):
        s = Shell(["exit 1"], "/dev/null")
        self.assertEqual(len(s.processes), 1)
        self.assertEqual(s.processes.itervalues().next()['command'], 'exit 1')
        self.assertEqual(s.processes.itervalues().next()['status'], 1)

    def test_signalled_command(self):
        s = Shell(["kill -s INT $$"], "/dev/null")
        self.assertEqual(len(s.processes), 1)
        self.assertEqual(s.processes.itervalues().next()['command'], 'kill -s INT $$')
        self.assertEqual(s.processes.itervalues().next()['status'], -2)

    def test_multiple_commands(self):
        s = Shell([
            "exit 0", "exit 1"
        ], "/dev/null")

        self.assertEqual(len(s.processes), 2)
        exit_statuses = {}
        for pid, process in s.processes.iteritems():
            if process["status"] not in exit_statuses:
                exit_statuses[process["status"]] = 0

            exit_statuses[process["status"]] += 1

        self.assertEqual(len(exit_statuses), 2)
        self.assertEqual(exit_statuses[0], 1)
        self.assertEqual(exit_statuses[1], 1)

    def test_command_output(self):
        # Delete 'output.txt' if present.
        try:
            os.remove('output.txt')
        except OSError:
            pass

        s = Shell([
            "echo foobar", "echo barbaz"
        ])

        # Check for presence of 'foobar' and 'barbaz'
        found_foobar = False
        found_barbaz = False
        with open("output.txt", 'r') as f:
            for line in f:
                if "foobar" in line:
                    found_foobar = True

                if "barbaz" in line:
                    found_barbaz = True

        self.assertTrue(found_foobar)
        self.assertTrue(found_barbaz)

        try:
            os.remove('output.txt')
        except OSError:
            pass

    def test_command_custom_output_file(self):
        # Delete 'foobar.txt' if present.
        try:
            os.remove('foobar.txt')
        except OSError:
            pass

        s = Shell([
            "echo foobar", "echo barbaz"
        ], "foobar.txt")

        # Check for presence of 'foobar' and 'barbaz'
        found_foobar = False
        found_barbaz = False
        with open("foobar.txt", 'r') as f:
            for line in f:
                if "foobar" in line:
                    found_foobar = True

                if "barbaz" in line:
                    found_barbaz = True

        self.assertTrue(found_foobar)
        self.assertTrue(found_barbaz)

        try:
            os.remove('foobar.txt')
        except OSError:
            pass
