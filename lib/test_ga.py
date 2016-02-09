import unittest
import ga
import os


class ExperimentTest(unittest.TestCase):
    def test_repetitions(self):
        class SimpleExperiment(ga.Experiment):
            def __init__(self):
                ga.Experiment.__init__(self)
                self.count = 0

                def count(configuration):
                    self.count += 1

                self.add_phase("count", count)

        s = SimpleExperiment()

        # Verify that default run count is 3
        s.run()
        self.assertEqual(s.count, 3)

        # Verify that another 3 runs makes 7 runs in total
        s.run(repetitions=4)
        self.assertEqual(s.count, 7)

    def test_permutations(self):
        class PermutationExperiment(ga.Experiment):
            def __init__(self):
                pass

        s = PermutationExperiment()
        self.assertEqual(s.generate_permutations(None), [None])
        self.assertEqual(s.generate_permutations([["A", "B"]]), [["A"], ["B"]])
        self.assertEqual(s.generate_permutations([["A", "B"], ["C", "D"]]),
                         [["A", "C"], ["A", "D"], ["B", "C"], ["B", "D"]])

    def test_directory_structure(self):
        class DirectoryExperiment(ga.Experiment):
            def __init__(self):
                ga.Experiment.__init__(self)

                def exp1(configuration):
                    pass

                def exp2(configuration):
                    pass

                def exp3(configuration):
                    pass

                self.add_phase("exp1", exp1)
                self.add_phase("exp2", exp2)
                self.add_phase("exp3", exp3, [["A", "B"], ["C", "D"]])

        s = DirectoryExperiment()
        s.run(2)

        self.assertNotEqual(s.run_id, None)

        self.assertTrue(os.path.exists("data/%s/exp1/run_0" % s.run_id))
        self.assertTrue(os.path.exists("data/%s/exp1/run_1" % s.run_id))
        self.assertTrue(os.path.exists("data/%s/exp2/run_0" % s.run_id))
        self.assertTrue(os.path.exists("data/%s/exp2/run_1" % s.run_id))
        self.assertTrue(os.path.exists("data/%s/exp3_A_C/run_0" % s.run_id))
        self.assertTrue(os.path.exists("data/%s/exp3_A_C/run_1" % s.run_id))
        self.assertTrue(os.path.exists("data/%s/exp3_A_D/run_0" % s.run_id))
        self.assertTrue(os.path.exists("data/%s/exp3_A_D/run_1" % s.run_id))
        self.assertTrue(os.path.exists("data/%s/exp3_B_C/run_0" % s.run_id))
        self.assertTrue(os.path.exists("data/%s/exp3_B_C/run_1" % s.run_id))
        self.assertTrue(os.path.exists("data/%s/exp3_B_D/run_0" % s.run_id))
        self.assertTrue(os.path.exists("data/%s/exp3_B_D/run_1" % s.run_id))
