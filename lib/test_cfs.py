import unittest
from cfs import CFS
import os


class CFSTest(unittest.TestCase):
    def test_other(self):
        self.assertEqual(str(CFS(CFS.SCHED_OTHER, 0, 'foo')),
                "chrt --other 0 foo")
        with self.assertRaises(Exception) as e:
            CFS(CFS.SCHED_OTHER, 5, 'foo')

    def test_idle(self):
        self.assertEqual(str(CFS(CFS.SCHED_IDLE, 0, 'foo')),
                "chrt --idle 0 foo")
        with self.assertRaises(Exception) as e:
            CFS(CFS.SCHED_IDLE, 5, 'foo')

    def test_batch(self):
        self.assertEqual(str(CFS(CFS.SCHED_BATCH, 0, 'foo')),
                "chrt --batch 0 foo")
        with self.assertRaises(Exception) as e:
            CFS(CFS.SCHED_BATCH, 5, 'foo')

    def test_round_robin(self):
        self.assertEqual(str(CFS(CFS.SCHED_RR, 1, 'foo')),
                "chrt --rr 1 foo")
        with self.assertRaises(Exception) as e:
            CFS(CFS.SCHED_RR, 0, 'foo')

    def test_fifo(self):
        self.assertEqual(str(CFS(CFS.SCHED_FIFO, 1, 'foo')),
                "chrt --fifo 1 foo")
        with self.assertRaises(Exception) as e:
            CFS(CFS.SCHED_FIFO, 0, 'foo')

if __name__ == '__main__':
    unittest.main()
