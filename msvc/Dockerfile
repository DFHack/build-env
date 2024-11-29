FROM ubuntu:22.04

LABEL org.opencontainers.image.source=https://github.com/DFHack/build-env

ARG BUILDER_UID=1001

# install dependencies and initialize the wine environment for root
RUN apt-get update \
 && apt-get upgrade -y \
 && apt-get install -y wine64 python3 msitools ca-certificates \
 && apt-get install -y git gcc-10 g++-10 ninja-build libxml-libxslt-perl make libssl-dev \
 && apt-get install -y dh-autoreconf libcurl4-gnutls-dev libexpat1-dev gettext libz-dev \
 && apt-get install -y wget curl unzip python3-pip \
 && apt-get clean -y \
 && rm -rf /var/lib/apt/lists/* \
 && pip3 install --compile sphinx \
 && pip3 install six \
 && useradd --uid $BUILDER_UID --create-home --shell /bin/bash buildmaster \
 && $(command -v wine64 || command -v wine || false) wineboot --init \
 && while pgrep wineserver > /dev/null; do sleep 1; done

# set up msvc
WORKDIR /opt/msvc
COPY lowercase fixinclude install.sh vsdownload.py msvctricks.cpp ./
COPY wrappers/* ./wrappers/

RUN PYTHONUNBUFFERED=1 ./vsdownload.py --accept-license --dest /opt/msvc \
 && ./install.sh /opt/msvc \
 && rm lowercase fixinclude install.sh vsdownload.py \
 && rm -rf wrappers

COPY msvcenv-native.sh /opt/msvc

# compile custom cmake for msvc and ninja compatibility
RUN mkdir -p /opt/cmake/src && cd /opt/cmake/src \
 && git clone -b msvc-3.22.1 --depth=1 https://gitlab.kitware.com/mstorsjo/cmake.git \
 && mkdir cmake/build && cd cmake/build && ../configure --prefix=/opt/cmake --parallel=$(nproc) \
 && make -j$(nproc) && make install \
 && cd && rm -rf /opt/cmake/src

# install newer ccache for msvc caching
RUN mkdir -p /opt/ccache && cd /opt/ccache \
 && wget 'https://github.com/ccache/ccache/releases/download/v4.8.3/ccache-4.8.3-linux-x86_64.tar.xz' -O- | tar -xJ \
 && mv * bin

ENV BIN=/opt/msvc/bin/x64 \
    PATH=/opt/msvc:/opt/msvc/bin/x64:/opt/ccache/bin:/opt/cmake/bin:$PATH \
    WINEARCH=win64 \
    WINEDLLOVERRIDES='ngen.exe,mscorsvw.exe=b' \
    WINEDEBUG=-all

# set up dfhack build environment
USER buildmaster
WORKDIR /home/buildmaster

# Add native build directory for DFHack cross compiling
RUN mkdir /home/buildmaster/dfhack-native \
 && cd /home/buildmaster/dfhack-native \
 && git clone --depth=1 https://github.com/DFHack/dfhack.git \
 && cd dfhack && git submodule update --init && cd .. \
 && CC=gcc-10 CXX=g++-10 cmake dfhack -GNinja -DCMAKE_BUILD_TYPE=Release \
 && ninja protoc-bin \
 && cd .. \
 && mkdir -p dfhack-native-bin/depends/protobuf \
 && mv dfhack-native/ImportExecutables.cmake dfhack-native-bin \
 && mv dfhack-native/depends/protobuf/*proto* dfhack-native-bin/depends/protobuf \
 && rm -rf dfhack-native \
 && mv dfhack-native-bin dfhack-native

# set up wine and ccache for the buildmaster user
RUN $(command -v wine64 || command -v wine || false) wineboot --init \
 && wineserver -w \
 && ccache -C

# copy in frequently-changing scripts last so the image can be quickly rebuilt
USER root
COPY dfhack-configure dfhack-make dfhack-test /usr/local/bin/

COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]

CMD ["bash"]
