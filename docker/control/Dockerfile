FROM golang:1.8
MAINTAINER siddontang@gmail.com

RUN apt-get -y -q update && \
    apt-get -y -q install software-properties-common && \
    apt-get -y -q update && \
    apt-get install -qqy \
        git \
        gnuplot \
        wget \
	vim # not required by chaos itself, just for ease of use

# You need to locate chaos in this directory (up.sh does that automatically)
ADD chaos /go/src/github.com/siddontang/chaos

ADD ./bashrc /root/.bashrc
ADD ./init.sh /init.sh
RUN chmod +x /init.sh

CMD /init.sh
