FROM buildpack-deps:bionic

RUN dpkg --add-architecture i386 \
 && apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        libxml-libxml-perl \
        libxml-libxslt-perl \
        lua5.3 \
        ninja-build \
        python3-pip \
        ruby \
        wine32 \
        winetricks \
        xvfb \
 && rm -rf /var/lib/apt/lists/* \
 && useradd --uid 1001 --create-home --shell /bin/bash buildmaster \
 && mkdir -p /opt/cmake \
 && cd /opt \
 && curl -LSo cmake.tar.gz https://github.com/Kitware/CMake/releases/download/v3.17.5/cmake-3.17.5-Linux-x86_64.tar.gz \
 && echo "897142368b15a5693c999a7ed2187be20c1b41a68c3711379d32a33469bb29ba  cmake.tar.gz" | sha256sum -c \
 && tar xzCf /opt/cmake /opt/cmake.tar.gz --strip-components=1 \
 && sed -e 's#/Zi#/Z7#g' -i /opt/cmake/share/cmake-3.22/Modules/Platform/Windows-MSVC.cmake \
 && rm -f /opt/cmake.tar.gz

RUN pip3 install --compile sphinx

USER buildmaster
WORKDIR /home/buildmaster

ENV PATH=/opt/cmake/bin:$PATH \
    WINEARCH=win32 \
    WINEDLLOVERRIDES='ngen.exe,mscorsvw.exe=b'

# Get rid of the "install mono?" and "install gecko?" prompts.
RUN WINEDLLOVERRIDES="mscoree,mshtml=" wine wineboot && wineserver -w

# Install dependencies.
RUN Xvfb :0 -screen 0 1024x768x24 \
  & LOGNAME=buildmaster DISPLAY=:0 winetricks --unattended \
        dotnet452 \
        vcrun2015 \
 && rm -rf /home/buildmaster/.cache/winetricks

# Add native build directory.
# (maintainer: Update cloned git tag only if needed.)
RUN mkdir /home/buildmaster/dfhack-native \
 && cd /home/buildmaster/dfhack-native \
 && git clone -b 0.44.12-r3 --depth=1 --recursive https://github.com/DFHack/dfhack.git \
 && cmake dfhack \
        -DDFHACK_BUILD_ARCH=64 \
        -DBUILD_LIBRARY=OFF \
        -DBUILD_PLUGINS=OFF \
        -DBUILD_RUBY=OFF \
        -DBUILD_SUPPORTED=OFF \
        -DDOWNLOAD_RUBY=OFF \
 && make -j$(nproc) protoc-bin \
 && cd .. \
 && mkdir -p dfhack-native-bin/depends/protobuf \
 && mv dfhack-native/ImportExecutables.cmake dfhack-native-bin \
 && mv dfhack-native/depends/protobuf/libprotobuf.so \
       dfhack-native/depends/protobuf/libprotoc.so \
       dfhack-native/depends/protobuf/protoc \
       dfhack-native-bin/depends/protobuf \
 && rm -rf dfhack-native \
 && mv dfhack-native-bin dfhack-native

# Set up wine_vcpp
# Based on https://github.com/eruffaldi/wine_vcpp
ADD ["Microsoft Visual Studio 14.0/VC","/home/buildmaster/wine_vcpp/VC"]
ADD ["Windows Kits/8.1","/home/buildmaster/wine_vcpp/kit8.1"]
ADD ["Windows Kits/10","/home/buildmaster/wine_vcpp/kit10"]
ADD ["Microsoft Visual Studio 14.0/Common7/Tools","/home/buildmaster/wine_vcpp/tools"]
RUN ln -s /home/buildmaster/wine_vcpp /home/buildmaster/.wine/drive_c/VC

# Add DFHack scripts.
USER root

RUN cd /root \
 && curl -LSo embed-resource.tar.gz https://github.com/CogentEmbedded/embed-resource/archive/b9b04474f76c8bc4b28260e6225845fc56e39a5d.tar.gz \
 && echo "489824c1d1d977fb0e2b116f7648dde9c5370c188dbf95b733c5c56677d7304b  embed-resource.tar.gz" | sha256sum -c \
 && tar xf embed-resource.tar.gz \
 && rm -f embed-resource.tar.gz \
 && mv embed-resource-b9b04474f76c8bc4b28260e6225845fc56e39a5d embed-resource \
 && cd embed-resource \
 && cmake -G Ninja -Wno-dev . \
 && ninja embed-resource \
 && mv embed-resource /usr/local/bin \
 && cd .. \
 && rm -rf embed-resource

COPY boost-1.67-precompiled.tar.xz /
RUN cd / \
 && tar xf boost-1.67-precompiled.tar.xz \
 && rm boost-1.67-precompiled.tar.xz

ADD dfhack-configure dfhack-make dfhack-test wrapper /usr/local/bin/
RUN ln -s wrapper /usr/local/bin/cl \
 && ln -s wrapper /usr/local/bin/link \
 && ln -s wrapper /usr/local/bin/mt \
 && ln -s wrapper /usr/local/bin/rc
USER buildmaster
ADD toolchain.cmake vcvars_32.bash vcvars_64.bash /home/buildmaster/

# Experimental: ccache with msvc support!
# Built from https://github.com/ccache/ccache/archive/3972b9bd6f9c27df02d277d7253b316ffc125995.tar.gz
ADD ccache.exe /home/buildmaster/.wine/drive_c/windows/system32/

ENV WINEDEBUG=-all
