FROM golang as gobuilder

RUN go get -d github.com/BenLubar/git-last-modified \
 && go build -o /git-last-modified github.com/BenLubar/git-last-modified

FROM buildpack-deps:bionic

ENV CCACHE_SLOPPINESS=file_stat_matches,file_stat_matches_ctime,include_file_ctime,include_file_mtime,locale \
    GCC_VERSION=7.5.0 \
    MACOSX_DEPLOYMENT_TARGET=10.6 \
    OSXCROSS_GCC_NO_STATIC_RUNTIME=1 \
    OSXCROSS_NO_INCLUDE_PATH_WARNINGS=1 \
    PATH=/usr/lib/ccache:/osxcross/target/bin:/opt/cmake/bin:$PATH

COPY --from=gobuilder /git-last-modified /usr/bin/
ADD osxcross-patches.diff /osxcross/osxcross-patches.diff

RUN dpkg --add-architecture i386 \
 && apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        clang \
        g++-multilib \
        gcc-multilib \
        libglu1-mesa \
        libgmp-dev \
        libgtk2.0-0 \
        libmpc-dev \
        libmpfr-dev \
        libopenal1 \
        libsdl-image1.2 \
        libsdl-ttf2.0-0 \
        libxml-libxml-perl \
        libxml-libxslt-perl \
        lua5.3 \
        mesa-common-dev \
        ninja-build \
        python3-pip \
        ruby \
        xvfb \
        zlib1g-dev:amd64 \
        zlib1g-dev:i386 \
 && rm -rf /var/lib/apt/lists/*

RUN pip3 install --compile sphinx

RUN mkdir -p /osxcross/tarballs /opt/cmake /usr/src/ccache \
 && cd /osxcross/tarballs \
 && curl -LSo cmake.tar.gz https://cmake.org/files/v3.11/cmake-3.11.4-Linux-x86_64.tar.gz \
 && curl -LSo osxcross.tar.gz https://github.com/tpoechtrager/osxcross/archive/1a1733a773fe26e7b6c93b16fbf9341f22fac831.tar.gz \
 && curl -LSo MacOSX10.10.sdk.tar.xz https://github.com/phracker/MacOSX-SDKs/releases/download/10.13/MacOSX10.10.sdk.tar.xz \
 && curl -LSo gcc-7.5.0.tar.gz https://ftpmirror.gnu.org/gcc/gcc-7.5.0/gcc-7.5.0.tar.gz \
 && curl -LSo ccache-3.7.9.tar.xz https://github.com/ccache/ccache/releases/download/v3.7.9/ccache-3.7.9.tar.xz \
 && (echo "6dab016a6b82082b8bcd0f4d1e53418d6372015dd983d29367b9153f1a376435  cmake.tar.gz"; \
     echo "c6cead036022edb7013a6adebf5c6832e06d5281b72515b10890bf91b8fe9ada  osxcross.tar.gz"; \
     echo "4a08de46b8e96f6db7ad3202054e28d7b3d60a3d38cd56e61f08fb4863c488ce  MacOSX10.10.sdk.tar.xz"; \
     echo "4f518f18cfb694ad7975064e99e200fe98af13603b47e67e801ba9580e50a07f  gcc-7.5.0.tar.gz"; \
     echo "f893da7543bfb9172bd55e603fcbdfcd83e6def176a28689c13235695b4cf44b  ccache-3.7.9.tar.xz") | sha256sum -c \
 && tar xzCf /opt/cmake /osxcross/tarballs/cmake.tar.gz --strip-components=1 \
 && tar xJCf /usr/src/ccache /osxcross/tarballs/ccache-3.7.9.tar.xz --strip-components=1 \
 && tar xzCf /osxcross /osxcross/tarballs/osxcross.tar.gz --strip-components=1 \
 && rm -f /osxcross/tarballs/cmake.tar.gz /osxcross/tarballs/ccache-3.7.9.tar.xz /osxcross/tarballs/osxcross.tar.gz

RUN cd /osxcross \
 && patch -p1 < osxcross-patches.diff \
 && UNATTENDED=1 ./build.sh
RUN cd /osxcross \
 && UNATTENDED=1 ./build_gcc.sh
RUN cd /osxcross \
 && UNATTENDED=1 ./build_llvm_dsymutil.sh
RUN cd /osxcross \
 && UNATTENDED=1 ./tools/osxcross-macports install zlib

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get purge -y --auto-remove \
        clang \
        libgmp-dev \
        libmpc-dev \
        libmpfr-dev \
 && rm -rf /var/lib/apt/lists/*

RUN ln -s /osxcross/target/macports/pkgs/opt/local/lib/libz.dylib /usr/lib/libz.dylib \
 && ln -s /bin/true /osxcross/target/bin/install_name_tool

RUN cd /usr/src/ccache \
 && ./configure --prefix=/usr \
 && make install \
 && rm -rf /usr/src/ccache

RUN useradd --uid 1001 --create-home --shell /bin/bash buildmaster \
 && mkdir /home/buildmaster/dfhack-native \
 && cd /home/buildmaster/dfhack-native \
 && git clone --depth=1 --recursive https://github.com/DFHack/dfhack.git \
 && cmake dfhack \
        -G Ninja \
        -DDFHACK_BUILD_ARCH=64 \
        -DBUILD_LIBRARY=OFF \
        -DBUILD_PLUGINS=OFF \
        -DBUILD_RUBY=OFF \
        -DBUILD_SUPPORTED=OFF \
        -DDOWNLOAD_RUBY=OFF \
 && ninja protoc-bin \
 && bash -e dfhack/travis/build-lua.sh \
 && chmod a+rx /root /root/lua53/bin/* \
 && chmod -R a+rX /root/lua53 \
 && cd /home/buildmaster \
 && mkdir -p dfhack-native-bin/depends/protobuf dfhack-native-bin/travis \
 && mv dfhack-native/ImportExecutables.cmake dfhack-native-bin \
 && mv dfhack-native/depends/protobuf/libprotobuf.so \
       dfhack-native/depends/protobuf/libprotoc.so \
       dfhack-native/depends/protobuf/protoc \
       dfhack-native-bin/depends/protobuf \
 && mv dfhack-native/dfhack/travis/run-tests.py dfhack-native-bin/travis \
 && rm -rf dfhack-native \
 && mv dfhack-native-bin dfhack-native

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        libsdl1.2debian:amd64 \
        libsdl1.2debian:i386 \
 && rm -rf /var/lib/apt/lists/* \
 && ln -s libSDL-1.2.so.0 /usr/lib/x86_64-linux-gnu/libSDL.so \
 && ln -s libSDL-1.2.so.0 /usr/lib/i386-linux-gnu/libSDL.so

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

RUN cd /root \
 && curl -LSo boost_1_67_0.tar.bz2 https://dl.bintray.com/boostorg/release/1.67.0/source/boost_1_67_0.tar.bz2 \
 && echo "2684c972994ee57fc5632e03bf044746f6eb45d4920c343937a465fd67a5adba  boost_1_67_0.tar.bz2" | sha256sum -c \
 && tar xf boost_1_67_0.tar.bz2 \
 && rm -f boost_1_67_0.tar.bz2 \
 && cd boost_1_67_0 \
 && ./bootstrap.sh \
 && ./b2 -d0 -j10 -q \
        --layout=versioned \
        --without-python \
        toolset=gcc \
        address-model=32,64 \
        cxxflags=-fPIC \
        variant=release \
        link=static \
        threading=multi \
        install \
 && ln -s /osxcross/target/bin/sw_vers /usr/bin/sw_vers \
 && ln -s /osxcross/target/bin/x86_64-apple-darwin14-libtool /usr/bin/libtool \
 && (echo 'using darwin : 7.5 : x86_64-apple-darwin14-g++ :'; \
     echo '<assembler>x86_64-apple-darwin14-as'; \
     echo '<ranlib>x86_64-apple-darwin14-ranlib ;') >> project-config.jam \
 && ./b2 -d0 -j10 -q \
        --layout=versioned \
        --without-python \
        toolset=darwin-7.5 \
        target-os=darwin \
        address-model=32,64 \
        cxxflags=-fPIC \
        variant=release \
        link=static \
        threading=multi \
        architecture=x86 \
        binary-format=mach-o \
        abi=sysv \
        install \
 && rm -rf /root/boost_1_67_0

ENV BOOST_LIBRARYDIR=/usr/local/lib

RUN mkdir -p /usr/lib/ccache \
 && ln -s ../../bin/ccache /usr/lib/ccache/g++ \
 && ln -s ../../bin/ccache /usr/lib/ccache/gcc

# Not in $PATH to avoid confusing build failures
RUN mkdir -p /usr/lib/ccache-osx \
 && ln -s ../../bin/ccache /usr/lib/ccache-osx/x86_64-apple-darwin14-g++ \
 && ln -s ../../bin/ccache /usr/lib/ccache-osx/x86_64-apple-darwin14-gcc

ADD dfhack-configure dfhack-make dfhack-test /usr/local/bin/

USER buildmaster
WORKDIR /home/buildmaster
