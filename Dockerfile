# Dockerfile example
FROM ubuntu:18.04
MAINTAINER Leo Lu (Lu Yu Xuan) <i@leoleoasd.me>

# install go
ENV PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUN apt-get update
RUN apt-get install -y clang-10
RUN update-alternatives --install /usr/bin/clang clang /usr/bin/clang-10 10000
RUN update-alternatives --install /usr/bin/clang++ clang++ /usr/bin/clang++-10 10000
RUN update-alternatives --install /usr/bin/llvm-config llvm-config /usr/bin/llvm-config-10 10000
RUN apt-get install -y python3.8
RUN update-alternatives --install /usr/bin/python python /usr/bin/python3.8 10000
RUN apt-get install -y wget
RUN wget https://golang.org/dl/go1.16.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.16.linux-amd64.tar.gz
RUN go version
RUN rm go1.16.linux-amd64.tar.gz

RUN apt-get install -y pkg-config gcc libseccomp-dev
RUN apt-get install -y unzip diffutils

RUN mkdir /judger
COPY . /judger
RUN cd /judger && go mod download && go build .
RUN useradd build_user
RUN useradd run_user

RUN mkdir -p /data/scripts
RUN mkdir -p /data/test_cases

CMD /judger/judgeServer
