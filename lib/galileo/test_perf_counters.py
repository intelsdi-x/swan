import unittest
from perf_counters import Perf, Timeline


class PerfTest(unittest.TestCase):
    def test_empty_metrics(self):
        self.assertEqual(str(Perf(events=None, command="foobar")),
                         "perf stat -x ',' --append -I 1000 -o perf.txt foobar")

    def test_empty_interval(self):
        self.assertEqual(str(Perf(interval=None, command="foobar")), "perf stat -x ',' --append -o perf.txt foobar")

    def test_one_metric(self):
        self.assertEqual(str(Perf(events=["A"], command="foobar")),
                         "perf stat -x ',' --append -e A -I 1000 -o perf.txt foobar")

    def test_multiple_metrics(self):
        self.assertEqual(str(Perf(events=["A", "B"], command="foobar")),
                         "perf stat -x ',' --append -e A,B -I 1000 -o perf.txt foobar")

    def test_custom_interval(self):
        self.assertEqual(str(Perf(events=["A", "B"], command="foobar", interval=2000)),
                         "perf stat -x ',' --append -e A,B -I 2000 -o perf.txt foobar")

    def test_custom_output(self):
        self.assertEqual(str(Perf(events=["A", "B"], command="foobar", interval=2000, output_file="barbaz.txt")),
                         "perf stat -x ',' --append -e A,B -I 2000 -o barbaz.txt foobar")

    def test_timeline_parse_empty_file(self):
        tl = Timeline("perf_examples/empty.txt")
        self.assertEqual(len(tl.entries), 0)

    def test_timeline_parse_single_metric(self):
        tl = Timeline("perf_examples/single_metric.txt")
        self.assertEqual(len(tl.entries), 10)

        previous_timestamp = 0
        for entry in tl.entries:
            self.assertGreater(entry.time, previous_timestamp)
            previous_timestamp = entry.time
            self.assertNotEqual(entry.time, None)
            self.assertEqual(len(entry.data), 1)

    def test_timeline_parse_multiple_metrics(self):
        tl = Timeline("perf_examples/two_metrics.txt")
        self.assertEqual(len(tl.entries), 10)

        previous_timestamp = 0
        for entry in tl.entries:
            self.assertGreater(entry.time, previous_timestamp)
            previous_timestamp = entry.time
            self.assertNotEqual(entry.time, None)
            self.assertEqual(len(entry.data), 2)

    def test_timeline_filter_single_metric(self):
        tl = Timeline("perf_examples/single_metric.txt")
        rows = tl.filter_by_columns(["time", "page-faults"])
        self.assertEqual(len(rows), 10)
        for row in rows:
            self.assertEqual(len(row), 2)

        rows = tl.filter_by_columns(["page-faults"])
        self.assertEqual(len(rows), 10)
        for row in rows:
            self.assertEqual(len(row), 1)

    def test_timeline_filter_multiple_metrics(self):
        tl = Timeline("perf_examples/two_metrics.txt")
        rows = tl.filter_by_columns(["time", "context-switches", "page-faults"])
        self.assertEqual(len(rows), 10)
        for row in rows:
            self.assertEqual(len(row), 3)

        rows = tl.filter_by_columns(["time", "context-switches"])
        self.assertEqual(len(rows), 10)
        for row in rows:
            self.assertEqual(len(row), 2)

        rows = tl.filter_by_columns(["context-switches"])
        self.assertEqual(len(rows), 10)
        for row in rows:
            self.assertEqual(len(row), 1)

    def test_timeline_filter_in_columns(self):
        tl = Timeline("perf_examples/two_metrics.txt")
        rows = tl.filter_by_columns(["time", "context-switches", "page-faults"], separate_columns=True)
        self.assertEqual(len(rows), 3)
        for row in rows:
            self.assertEqual(len(row), 10)

        rows = tl.filter_by_columns(["time", "context-switches"], separate_columns=True)
        self.assertEqual(len(rows), 2)
        for row in rows:
            self.assertEqual(len(row), 10)

        rows = tl.filter_by_columns(["context-switches"], separate_columns=True)
        self.assertEqual(len(rows), 1)
        for row in rows:
            self.assertEqual(len(row), 10)

    def test_timeline_tsv(self):
        tl = Timeline("perf_examples/two_metrics.txt")
        tsv = tl.tsv()

        # 10 data points + legend line
        self.assertEqual(len(tsv), 11)

        # Test legend, first and last data point
        self.assertEqual(tsv[0], "#time\tcontext-switches\tpage-faults")
        self.assertEqual(tsv[1], "1.001251192\t10.0\t427.0")
        self.assertEqual(tsv[9], "9.006753602\t4.0\t0.0")

    def test_timeline_csv(self):
        tl = Timeline("perf_examples/two_metrics.txt")
        csv = tl.csv()

        # 10 data points + legend line
        self.assertEqual(len(csv), 11)

        # Test legend, first and last data point
        self.assertEqual(csv[0], "#time,context-switches,page-faults")
        self.assertEqual(csv[1], "1.001251192,10.0,427.0")
        self.assertEqual(csv[9], "9.006753602,4.0,0.0")
