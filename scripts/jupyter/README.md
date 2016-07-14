# Jupyter experiment viewer

## Installation

You need `python` and `pip` to install the necessary dependencies for Jupyter.
On Centos 7, install the following packages with:

```sh
sudo yum install python-pip python-devel
```

After this, install the python dependencies with `pip` with:

```sh
sudo pip install -r requirements.txt
```

## Launching jupyter

Start Jupyter by running the following in the `swan/scripts/jupyter` directory:

```sh
jupyter notebook --ip=0.0.0.0
```

If run locally, the command will bring up the default browser.
If not, connect to http://hostname:8888/ through your browser.

## Explore data using Jupyter

From within the Jupyter interface, create a new notebook by clicking on 'New' and 'Python 2'.

![experiments list](docs/new_notebook.png)

Within the new notebook, import the experiments module by typing:

```
from experiments import *
```

And evaluate the expression by clicking shift and enter.
Then, connect to the Cassandra database and load the list of experiments:

```
experiments = Experiments(['address to Cassandra server'])
```

To load the available samples for an experiment, run:
```
experiment = experiments.experiment('uuid of experiment')
```

Showing the available samples can be done in the similar manner:

![sample list](docs/sample_list.png)

To render a sensitivity profile from the loaded samples, run:
```
profile = experiment.profile(slo=500)
```

Where 500 is the target latency in micro seconds.

This should render a table similar to the one below:

![sensitivity profile](docs/sensitivity_profile.png)
