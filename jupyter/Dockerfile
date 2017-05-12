FROM jupyter/minimal-notebook

RUN conda install -q -y nomkl cassandra-driver pandas pep8 pylint
RUN conda clean -y -a -v 

ADD test_data/ $HOME/work/test_data
ADD example.ipynb $HOME/work
ADD *.py $HOME/work/

RUN chown -R $USER:$USER $HOME
