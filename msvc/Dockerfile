FROM ubuntu:20.04

ARG BUILDER_UID=1001

# install dependencies
RUN apt-get update \
 && apt-get upgrade -y \
 && apt-get install -y wine64-development python3 msitools python3-simplejson python3-six ca-certificates \
 && apt-get install -y gcc g++ ninja-build git zlib1g-dev libsdl1.2-dev libxml-libxml-perl libxml-libxslt-perl make libncurses-dev libssl-dev \
 && apt-get install -y dh-autoreconf libcurl4-gnutls-dev libexpat1-dev gettext libz-dev \
 && apt-get install -y wget curl unzip vim bash-completion xvfb python3-pip \
 && apt-get clean -y \
 && rm -rf /var/lib/apt/lists/* \
 && pip3 install --compile sphinx \
 && useradd --uid $BUILDER_UID --create-home --shell /bin/bash buildmaster

# set up msvc
WORKDIR /opt/msvc
COPY lowercase fixinclude install.sh vsdownload.py msvcenv-native.sh ./
COPY wrappers/* ./wrappers/

# This removes the arm and x86 bits to save space (around 5.5G)
RUN PYTHONUNBUFFERED=1 ./vsdownload.py --accept-license --dest /opt/msvc \
 && ./install.sh /opt/msvc \
 && rm lowercase fixinclude install.sh vsdownload.py \
 && rm -rf wrappers \
 && find /opt/msvc -depth -type d -iregex '.*/.*arm[0-9]*$' -exec rm -fr {} \; \
 && find /opt/msvc -depth -type d -iregex '.*/.*x86$' -exec rm -fr {} \;

ENV BIN=/opt/msvc/bin/x64 \
    PATH=/opt/msvc:/opt/msvc/bin/x64:/opt/git/bin:/opt/ccache/bin:/opt/cmake/bin:$PATH \
    WINEARCH=win64 \
    WINEDLLOVERRIDES='ngen.exe,mscorsvw.exe=b' \
    WINEDEBUG=-all

# compile more recent git for ssh keysigning
RUN mkdir -p /opt/git/src && cd /opt/git/src \
 && wget https://github.com/git/git/archive/refs/tags/v2.38.1.tar.gz \
 && tar xzf "v2.38.1.tar.gz" \
 && cd git-2.38.1 \
 && make configure && ./configure --prefix=/opt/git \
 && make -j$(nproc) all && make install \
 && cd && rm -rf /opt/git/src

# compile custom cmake for msvc and ninja compatibility
RUN mkdir -p /opt/cmake/src && cd /opt/cmake/src \
 && git clone -b msvc-3.22.1 --depth=1 https://gitlab.kitware.com/mstorsjo/cmake.git \
 && mkdir cmake/build && cd cmake/build && ../configure --prefix=/opt/cmake --parallel=$(nproc) -- -DBUILD_CursesDialog=ON \
 && make -j$(nproc) && make install \
 && cd && rm -rf /opt/cmake/src

# compile custom ccache for msvc caching
COPY ccache-4.7.4.patch /tmp
RUN mkdir -p /opt/ccache/src && cd /opt/ccache/src \
 && git clone -b "v4.7.4" --depth=1 https://github.com/ccache/ccache.git \
 && mkdir ccache/build && cd ccache && git apply /tmp/ccache-4.7.4.patch \
 && cd build && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=/opt/ccache .. \
 && make -j$(nproc) && make install && cd && rm -rf /opt/ccache/src

# set up dfhack build environment
USER buildmaster
WORKDIR /home/buildmaster

# Add native build directory for DFHack cross compiling
# maintainer: Update cloned git tag when we update protobuf to keep them in sync
RUN mkdir /home/buildmaster/dfhack-native \
 && cd /home/buildmaster/dfhack-native \
 && git clone -b "0.47.05-r8" --depth=1 https://github.com/DFHack/dfhack.git \
 && cd dfhack && git submodule update --init && cd .. \
 && cmake dfhack -GNinja -DCMAKE_BUILD_TYPE=Release -DDOWNLOAD_RUBY=OFF \
 && ninja protoc-bin \
 && cd .. \
 && mkdir -p dfhack-native-bin/depends/protobuf \
 && mv dfhack-native/ImportExecutables.cmake dfhack-native-bin \
 && mv dfhack-native/depends/protobuf/*proto* dfhack-native-bin/depends/protobuf \
 && rm -rf dfhack-native \
 && mv dfhack-native-bin dfhack-native

# set up wine (get rid of the "install mono?" and "install gecko?" prompts) and ccache
RUN WINEDLLOVERRIDES="mscoree,mshtml=" wine64 wineboot --init \
 && wineserver -w \
 && ccache -C

# copy in frequently-changing scripts last so the image can be quickly rebuilt
USER root
COPY dfhack-configure dfhack-make dfhack-test /usr/local/bin/

COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]

CMD bash
