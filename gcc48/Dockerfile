FROM buildpack-deps:trusty AS osxcross-48

ENV MACOSX_DEPLOYMENT_TARGET=10.6 \
    OSXCROSS_GCC_NO_STATIC_RUNTIME=1

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        clang \
        cmake3 \
        libgmp-dev \
        libmpfr-dev \
        libmpc-dev \
 && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /osxcross/tarballs \
 && cd /osxcross/tarballs \
 && curl -LSo osxcross.tar.gz https://github.com/tpoechtrager/osxcross/archive/1a1733a773fe26e7b6c93b16fbf9341f22fac831.tar.gz \
 && curl -LSo MacOSX10.10.sdk.tar.xz https://github.com/phracker/MacOSX-SDKs/releases/download/10.13/MacOSX10.10.sdk.tar.xz \
 && (echo "c6cead036022edb7013a6adebf5c6832e06d5281b72515b10890bf91b8fe9ada  osxcross.tar.gz"; \
     echo "4a08de46b8e96f6db7ad3202054e28d7b3d60a3d38cd56e61f08fb4863c488ce  MacOSX10.10.sdk.tar.xz") | sha256sum -c

ADD osxcross-patches.diff /osxcross/osxcross-patches.diff

RUN tar xzCf /osxcross /osxcross/tarballs/osxcross.tar.gz --strip-components=1 \
 && cd /osxcross \
 && patch -p1 < osxcross-patches.diff

RUN cd /osxcross/tarballs \
 && curl -LSo gcc-4.8.5.tar.gz https://ftpmirror.gnu.org/gcc/gcc-4.8.5/gcc-4.8.5.tar.gz \
 && echo "1dbc5cd94c9947fe5dffd298e569de7f44c3cedbd428fceea59490d336d8295a  gcc-4.8.5.tar.gz" | sha256sum -c

ENV PATH=/osxcross48/target/bin:$PATH

RUN ln -s /osxcross /osxcross48 \
 && cd /osxcross48 \
 && UNATTENDED=1 ./build.sh \
 && UNATTENDED=1 GCC_VERSION=4.8.5 ./build_gcc.sh \
 && UNATTENDED=1 ./build_llvm_dsymutil.sh \
 && UNATTENDED=1 ./tools/osxcross-macports install zlib

FROM buildpack-deps:trusty

ENV MACOSX_DEPLOYMENT_TARGET=10.6 \
    OSXCROSS_GCC_NO_STATIC_RUNTIME=1 \
    PATH=/usr/lib/ccache:/osxcross48/target/bin:$PATH \
    CCACHE_SLOPPINESS=file_macro,include_file_ctime,include_file_mtime

RUN dpkg --add-architecture i386 \
 && apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ccache \
        cmake3 \
        g++-4.8-multilib \
        gcc-4.8-multilib \
        libxml-libxml-perl \
        libxml-libxslt-perl \
        mesa-common-dev \
        protobuf-compiler \
        python-pip \
        zlib1g-dev:amd64 \
        zlib1g-dev:i386 \
 && rm -rf /var/lib/apt/lists/* \
 && pip install sphinx

COPY --from=osxcross-48 /osxcross48/target /osxcross48/target
ADD dfhack-configure /usr/local/bin/dfhack-configure
ADD dfhack-make /usr/local/bin/dfhack-make

RUN ln -s ../../bin/ccache /usr/lib/ccache/x86_64-apple-darwin14-gcc \
 && ln -s ../../bin/ccache /usr/lib/ccache/x86_64-apple-darwin14-g++ \
 && ln -s /osxcross48/target/macports/pkgs/opt/local/lib/libz.dylib /usr/lib/libz.dylib \
 && ln -s /bin/true /osxcross48/target/bin/install_name_tool \
 && useradd --uid 1001 --create-home --shell /bin/bash buildmaster \
 && mkdir /home/buildmaster/dfhack-native \
 && (echo 'add_executable(protoc-bin IMPORTED)'; \
     echo 'set_property(TARGET protoc-bin APPEND PROPERTY IMPORTED_CONFIGURATIONS RELWITHDEBINFO;RELEASE)'; \
     echo 'set_target_properties(protoc-bin PROPERTIES IMPORTED_LOCATION_RELWITHDEBINFO "/usr/bin/protoc" IMPORTED_LOCATION_RELEASE "/usr/bin/protoc")') > /home/buildmaster/dfhack-native/ImportExecutables.cmake
