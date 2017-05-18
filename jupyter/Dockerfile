FROM jupyter/minimal-notebook

RUN conda install -q -y nomkl cassandra-driver pandas pep8 pylint
RUN conda clean -y -a -v 

COPY test_data example.ipynb *.py $HOME/work/

RUN chown -v -R $USER:$USER $HOME/work
