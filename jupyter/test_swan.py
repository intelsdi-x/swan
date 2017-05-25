"""
 Copyright (c) 2017 Intel Corporation

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
"""

import unittest
import swan
import shutil
import os


class TestExperiments(unittest.TestCase):

    def setUp(self):
        # prepare cached used by swan internally as mock
        try:
            os.makedirs(swan.DataFrameToCSVCache.CACHE_DIR)
        except OSError:
            # ignore dir exits
            pass

        for fn in ['6ed71a63-14c0-8e3e-3a65-4da9a11eecf6.csv.bz2',  # sensitivity profile
                   '80ad81ec-e6d7-cfc2-de6c-6c60cb300d7f.csv.bz2',  # optimal core allocation
                   'bc1ee530-4e02-b9fd-e845-752eb7545773.csv.bz2'   # memcached CAT
                   ]:
            src = os.path.join('test_data', fn)
            dst = os.path.join(swan.DataFrameToCSVCache.CACHE_DIR, fn)
            shutil.copyfile(src, dst)

    def assertRenders(self, styler):
        self.assertIsNotNone(
            styler._repr_html_()
        )

    def test_sensitivity_profile(self):
        profile = swan.SensitivityProfile('6ed71a63-14c0-8e3e-3a65-4da9a11eecf6', slo=500)
        self.assertRenders(profile.latency())
        self.assertRenders(profile.latency(normalized=False))
        self.assertRenders(profile.qps())
        self.assertRenders(profile.qps(normalized=False))
        self.assertRenders(profile.caffe_batches())

    def test_optimal_core_allocation(self):
        core = swan.OptimalCoreAllocation('80ad81ec-e6d7-cfc2-de6c-6c60cb300d7f', slo=500)
        self.assertRenders(core.latency())
        self.assertRenders(core.latency(normalized=False))
        self.assertRenders(core.qps())
        self.assertRenders(core.qps(normalized=False))
        self.assertRenders(core.cpu())

    def test_cat(self):
        core = swan.CAT('bc1ee530-4e02-b9fd-e845-752eb7545773', slo=500)
        self.assertRenders(core.latency())
        self.assertRenders(core.latency(normalized=False))
        self.assertRenders(core.latency(aggressor='Caffe', qps=500000))

        core.filtered_df()
        core.filtered_df_table()


if __name__ == '__main__':
    unittest.main()
