#FROM golang AS msvc-downloader
#
#ENV GOOS=windows GOARCH=amd64
#
#COPY download-msvc.go /go/src/github.com/BenLubar/build-env/download-msvc.go
#
#RUN cd /go/src/github.com/BenLubar/build-env \
# && git init \
# && git remote add origin https://github.com/BenLubar/build-env.git \
# && go get -d -v . \
# && go build -o /download-msvc.exe
#
#FROM buildpack-deps:bionic AS msvc
#
#RUN dpkg --add-architecture i386 \
# && apt-get update \
# && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
#        wine-development \
#        wine32-development \
#        wine64-development \
#        winetricks \
#        xvfb \
# && rm -rf /var/lib/apt/lists/* \
# && useradd --uid 1001 --create-home --shell /bin/bash buildmaster
#
#USER buildmaster
#WORKDIR /home/buildmaster
#
#RUN LOGNAME=buildmaster WINEDEBUG=-all xvfb-run -a winetricks --unattended \
#        dotnet462 \
#        vcrun2017 \
#	win10 \
# && rm -rf /home/buildmaster/.cache/winetricks
#
#COPY --from=msvc-downloader /download-msvc.exe /home/buildmaster/download-msvc.exe
#
#RUN wine download-msvc.exe
#
#RUN curl -LSo clcache.zip https://github.com/frerich/clcache/releases/download/v4.1.0/clcache-4.1.0.zip \
# && echo "3d41d98000e1edb705f2237204b943edd95afeb95a6f33c9ff87aa986471cce7  clcache.zip" | sha256sum -c \
# && unzip clcache.zip -d /home/buildmaster/.wine/drive_c/clcache \
# && rm -f clcache.zip

FROM buildpack-deps:bionic

ENV CCACHE_SLOPPINESS=file_macro,include_file_ctime,include_file_mtime \
    GCC_VERSION=7.3.0 \
    MACOSX_DEPLOYMENT_TARGET=10.6 \
    OSXCROSS_GCC_NO_STATIC_RUNTIME=1 \
    PATH=/usr/lib/ccache:/osxcross/target/bin:/opt/cmake/bin:$PATH \
    WINEDEBUG=-all,err+all

ADD osxcross-patches.diff /osxcross/osxcross-patches.diff

RUN dpkg --add-architecture i386 \
 && apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        clang \
        g++-multilib \
        gcc-multilib \
        libgmp-dev \
        libmpc-dev \
        libmpfr-dev \
        libxml-libxml-perl \
        libxml-libxslt-perl \
        mesa-common-dev \
        python3-sphinx \
        zlib1g-dev:amd64 \
        zlib1g-dev:i386 \
 && mkdir -p /osxcross/tarballs /opt/cmake \
 && cd /osxcross/tarballs \
 && curl -LSo cmake.tar.gz https://cmake.org/files/v3.11/cmake-3.11.1-Linux-x86_64.tar.gz \
 && curl -LSo osxcross.tar.gz https://github.com/tpoechtrager/osxcross/archive/1a1733a773fe26e7b6c93b16fbf9341f22fac831.tar.gz \
 && curl -LSo MacOSX10.10.sdk.tar.xz https://github.com/phracker/MacOSX-SDKs/releases/download/10.13/MacOSX10.10.sdk.tar.xz \
 && curl -LSo gcc-7.3.0.tar.gz https://ftpmirror.gnu.org/gcc/gcc-7.3.0/gcc-7.3.0.tar.gz \
 && (echo "3aa8d3b53903563bdc13dfec80716c286b2228db36ef65c23bc616f9bb930367  cmake.tar.gz"; \
     echo "c6cead036022edb7013a6adebf5c6832e06d5281b72515b10890bf91b8fe9ada  osxcross.tar.gz"; \
     echo "4a08de46b8e96f6db7ad3202054e28d7b3d60a3d38cd56e61f08fb4863c488ce  MacOSX10.10.sdk.tar.xz"; \
     echo "fa06e455ca198ddc11ea4ddf2a394cf7cfb66aa7e0ab98cc1184189f1d405870  gcc-7.3.0.tar.gz") | sha256sum -c \
 && tar xzCf /opt/cmake /osxcross/tarballs/cmake.tar.gz --strip-components=1 \
 && tar xzCf /osxcross /osxcross/tarballs/osxcross.tar.gz --strip-components=1 \
 && cd /osxcross \
 && patch -p1 < osxcross-patches.diff \
 && UNATTENDED=1 ./build.sh \
 && UNATTENDED=1 ./build_gcc.sh \
 && UNATTENDED=1 ./build_llvm_dsymutil.sh \
 && UNATTENDED=1 ./tools/osxcross-macports install zlib \
 && mv /osxcross/target /osxcross-target \
 && rm -rf /osxcross \
 && mkdir /osxcross \
 && mv /osxcross-target /osxcross/target \
 && DEBIAN_FRONTEND=noninteractive apt-get purge -y --auto-remove \
        clang \
        libgmp-dev \
        libmpc-dev \
        libmpfr-dev \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ccache \
 && rm -rf /var/lib/apt/lists/* \
 && ln -s /osxcross/target/macports/pkgs/opt/local/lib/libz.dylib /usr/lib/libz.dylib \
 && ln -s /bin/true /osxcross/target/bin/install_name_tool \
 && useradd --uid 1001 --create-home --shell /bin/bash buildmaster \
 && mkdir /home/buildmaster/dfhack-native \
 && cd /home/buildmaster/dfhack-native \
 && git clone --depth=1 https://github.com/DFHack/dfhack.git \
 && cmake dfhack/depends/protobuf \
 && make -j$(nproc) \
 && rm -rf dfhack \
 && chmod -R buildmaster:buildmaster /home/buildmaster/dfhack-native

ADD dfhack-configure dfhack-make /usr/local/bin/

USER buildmaster
WORKDIR /home/buildmaster
