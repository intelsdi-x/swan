<!--
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
-->

# Jupyter experiment viewer

## Introduction

> "The Jupyter Notebook is a web application that allows you to create and share documents that contain live code, equations, visualizations and explanatory text. Uses include: data cleaning and transformation, numerical simulation, statistical modeling, machine learning and much more." [from jupyter.org](http://jupyter.org/)

Swan uses *Jupyter Notebook* to filter, process and visualize results from experiments.

You need to have Docker installed to run Jupyter notebooks easily. We have provided a Docker image that will make running notebooks as easy as possible. You just need to run:

```sh
docker run -p 127.0.0.1:8888:8888 intelsdi/swan-jupyter
```

## Explore the Example Jupyter Notebook

From within the Jupyter web interface, open a template notebook by clicking on `example.ipynb`.

![notebook tree](/images/jupter-tree.png)

This is very simple notebook that will generate only sensitivity profile for the experiment.
The first step is to set the following variables:
- `EXPERIMENT_ID` is the identifier of the experiment which will be examined

After filling the variables, navigate to the green box using the keyboard arrows so that it points to the first variable and press `[Shift] [Enter]` to evaluate it. Evaluation actually means executing the code in the box. Evaluate further and observe the output. `Experiment` object's construction will look like:

```python
# An experiment can now be loaded from the database by its ID.
from swan import Experiment, SensitivityProfile
profile1 = SensitivityProfile(EXPERIMENT_ID, slo=500)
```
It may take a while since it will retrieve data from Cassandra and store it in the variable `profile1`.

The next step is to render the sensitivity profile from the loaded samples and draw sensitivity chart. The former will be generated after evaluating:

```python
profile1.latency()
```

resulting with:

![sample profile table](/images/jupyter-profile1-table.png)

To learn more about *Sensitivity Profile* read the [Sensitivity Experiment](experiments/memcached-sensitivity-profile/README.md) README.
