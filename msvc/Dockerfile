FROM ubuntu:20.04

RUN apt-get update \
 && apt-get upgrade -y \
 && apt-get install -y wine64-development python3 msitools python3-simplejson python3-six ca-certificates \
 && apt-get install -y gcc g++ ninja-build git zlib1g-dev libsdl1.2-dev libxml-libxml-perl libxml-libxslt-perl make libncurses-dev libssl-dev \
 && mkdir -p /opt/cmake/src && cd /opt/cmake/src \
 && git clone -b "msvc-3.22.1" --depth=1 https://gitlab.kitware.com/mstorsjo/cmake.git \
 && mkdir cmake/build && cd cmake/build && ../configure --prefix=/opt/cmake --parallel=$(nproc) -- -DBUILD_CursesDialog=ON \
 && make -j$(nproc) && make install && cd && rm -rf /opt/cmake/src \
 && apt-get install -y wget curl unzip vim bash-completion ccache xvfb python3-pip \
 && apt-get clean -y \
 && rm -rf /var/lib/apt/lists/* \
 && useradd --uid 1001 --create-home --shell /bin/bash buildmaster

RUN pip3 install --compile sphinx

WORKDIR /opt/msvc
COPY lowercase fixinclude install.sh vsdownload.py ./
COPY wrappers/* ./wrappers/

RUN PYTHONUNBUFFERED=1 ./vsdownload.py --accept-license --dest /opt/msvc && \
    ./install.sh /opt/msvc && \
    rm lowercase fixinclude install.sh vsdownload.py && \
    rm -rf wrappers

COPY msvcenv-native.sh /opt/msvc

USER buildmaster
WORKDIR /home/buildmaster

ENV BIN=/opt/msvc/bin/x64 \
    PATH=/opt/msvc:/opt/msvc/bin/x64:/opt/cmake/bin:$PATH \
    WINEARCH=win64 \
    WINEDLLOVERRIDES='ngen.exe,mscorsvw.exe=b' \
    WINEDEBUG=-all

# Add native build directory.
# (maintainer: Update cloned git tag only if needed.)
RUN mkdir /home/buildmaster/dfhack-native \
 && cd /home/buildmaster/dfhack-native \
 && git clone -n https://github.com/DFHack/dfhack.git \
 && cd dfhack && git checkout 9777f8ba029e47c747a034fc281e13ca70067c09 \
 && git submodule update --init \
 && cd build \
 && cmake .. -GNinja -DCMAKE_BUILD_TYPE=Release -DDOWNLOAD_RUBY=OFF && ninja protoc \
 && cd ../.. \
 && mkdir -p depends/protobuf \
 && mv dfhack/build/ImportExecutables.cmake . \
 && cp dfhack/build/depends/protobuf/protoc depends/protobuf \
 && rm -rf dfhack

# Get rid of the "install mono?" and "install gecko?" prompts.
RUN WINEDLLOVERRIDES="mscoree,mshtml=" wine64 wineboot --init && wineserver -w

# Experimental: ccache with msvc support!
# Built from https://github.com/ccache/ccache/archive/3972b9bd6f9c27df02d277d7253b316ffc125995.tar.gz
ADD ccache.exe /home/buildmaster/.wine/drive_c/windows/system32/

USER root
ADD dfhack-configure dfhack-make dfhack-test /usr/local/bin/
