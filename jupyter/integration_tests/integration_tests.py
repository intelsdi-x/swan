
# coding: utf-8

# In[ ]:

import sys
import unittest

import numpy as np

from pandas.util.testing import assert_frame_equal


# In[ ]:

from jupyter.experiment import Experiment
from jupyter.profile import Profile


# In[ ]:

EXPERIMENT_ID_SWAN = sys.stdin.readline().strip()


# In[ ]:

IP = '127.0.0.1'
PORT = 9042


# In[ ]:

exp1 = Experiment(cassandra_cluster=[IP], experiment_id=EXPERIMENT_ID_SWAN, port=PORT, name="experiment 1",
                  dir_csv='../test_data')


# In[ ]:

p1 = Profile(exp1, slo=500)
p1.sensitivity_table(show_throughput=True)


# In[ ]:

class TestExperiment(unittest.TestCase):
    def test_name_and_cached(self):
        self.assertEqual(exp1.name, 'experiment 1')
        self.assertEqual(exp1.cached_experiment, '../test_data/%s.csv' % EXPERIMENT_ID_SWAN)

    def test_cassandra_state_and_data_from_it(self):
        self.assertNotIsInstance(exp1.CASSANDRA_SESSION, None)
        self.assertNotIsInstance(exp1.get_frame(), None)


# In[ ]:

class TestProfile(unittest.TestCase):
    def test_exp_and_df_are_still_the_same(self):
        self.assertEqual(id(exp1), id(p1.exp))
        self.assertEqual(id(exp1.get_frame()), id(p1.data_frame))

    def test_sensitivity_table_and_charts_params(self):
        self.assertEqual(p1.slo, 500)
        self.assertEqual(p1.categories, ['Baseline', 'Caffe'])
        self.assertEqual(p1.latency_qps_aggrs.keys(), ['x', 'slo', 'Caffe', 'Baseline'])


# In[ ]:

class TestDataFrame(unittest.TestCase):
    def test_dimensions(self):
        self.assertTrue(exp1.get_frame().equals(p1.data_frame))
        self.assertEqual(exp1.frame['ns'].count(), 180)

    def test_some_numpy_data_arrays(self):
        self.assertTrue(np.array_equal(p1.frame.columns.get_values(), np.array([10., 20., 30., 40., 50., 60., 70., 80., 90., 100.])))
        self.assertTrue(np.alltrue(p1.frame.index.values, np.array(['Baseline', 'Caffe'])))


# In[ ]:

if __name__ == '__main__':
    unittest.main()
