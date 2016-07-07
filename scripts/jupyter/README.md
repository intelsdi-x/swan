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
jupyter notebook --ip=111.222.333.444
```
**NOTE** Remember to replace `111.222.333.444` with the IP of the host.

From that point on, you should be able to navigate to [http://111.222.333.444:8888/](http://111.222.333.444:8888/)

## Explore data using Jupyter
