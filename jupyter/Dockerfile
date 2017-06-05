FROM jupyter/minimal-notebook

RUN conda install -q -y nomkl cassandra-driver pandas pep8 pylint \
    && conda clean -y -a -v
RUN pip install pivottablejs

COPY test_data/ $HOME/work/test_data
COPY example.ipynb *.py $HOME/work/
COPY NOTICE $HOME/work/

# Switch to root and fix permissions
USER root
RUN chown -v -R $NB_USER:users /home/$NB_USER/work

# Switch back to jovyan user
USER $NB_USER
