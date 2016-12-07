# Jupyter experiment viewer

## Installation

You need `python` and `pip` to install the necessary dependencies for Jupyter.
On Centos 7, install the following packages with:

```sh
sudo yum install python-pip python-devel
```
or follow the instructions at [official pip site](https://pip.pypa.io/en/stable/installing/#installing-with-get-pip-py)

After this, install the python dependencies with `make` with:

```sh
make deps_jupyter
```

## Launching jupyter

Start Jupyter by running the following in the `swan/scripts/jupyter` directory:

```sh
jupyter notebook --ip=0.0.0.0 --port=8080
```

If run locally, the command will bring up the default browser.
If not, connect to http://hostname:8080/ through your browser.

## Explore data using Jupyter

From within the Jupyter interface, open a template notebook by clicking on `Open` and `example.ipynb`, or you can open a new natebook like below:

![experiment](docs/new_notebook.png)

Within the open tamplate notebook:
- set the `IP`, `PORT` of cassandra cluster and `EXPERIMENT_ID`
- focus on first `import` python statement:

```python
from experiment import *
```

And evaluate the expressions by clicking `shift` and `enter` on the cell.
```python
exp = Experiment(cassandra_cluster=['localhost'], experiment_id='uuid of experiment', port=9042)
```

Code above shows the available samples. Be aware that if a experiments has large data, it can take a while:

![sample list](docs/sample_list.png) 

If you want to get [pandas](http://pandas.pydata.org/) DataFrame from `exp` for deeper analisys you can get it like: 
```python
df1 = exp1.get_frame()
```
To render a sensitivity profile from the loaded samples, run:
```python
p = Profile(exp, slo=500)
p.sensitivity_table(show_throughput=False)
```

Where `slo` is the target latency in micro seconds and `show_throughput` is optional parameter and consist work make by aggressor.

This should render a table similar to the one below:

![sensitivity profile](docs/sensitivity_profile.png)

Below there are showed some missing data that can happen if you will try build sensivity profile in case of missing data in Cassandra.
In this case field in the table is marked as grey with `N/A`

```python
p1 = Profile(exp, slo=500)
p1.sensitivity_table(show_throughput=True)
```
![sensitivity profile](docs/sensitivity_profile_failed.png)

## Visualizing data using Jupyter

We are using [plotly](https://plot.ly/) interactive plots. There are some already prepared function for plots
the data directly in Jupyter, like:

```python
p1.sensitivity_chart(fill=True, to_max=True)
```
Where `fill` parameter fills area between Baseline and Aggressor. `to_max` shows comparison between Baseline and 'worst case'.

'worst case' in this case means max latency violations for all aggressors in each load point.