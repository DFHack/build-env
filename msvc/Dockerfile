FROM buildpack-deps:bionic AS wine_vcpp

# Based on https://github.com/eruffaldi/wine_vcpp
ADD ["Microsoft Visual Studio 14.0/VC","/home/buildmaster/wine_vcpp/VC"]
ADD ["Windows Kits/8.1","/home/buildmaster/wine_vcpp/kit8.1"]
ADD ["Windows Kits/10","/home/buildmaster/wine_vcpp/kit10"]
ADD ["Microsoft Visual Studio 14.0/Common7/Tools","/home/buildmaster/wine_vcpp/tools"]
#RUN find /home/buildmaster/wine_vcpp -type d -\( -name 'arm' -o -name 'arm64' -o -name 'amd64' -o -name 'amd64_arm' -o -name 'amd64_x86' -o -name 'x86_arm' -\) -exec rm -rf {} +

FROM buildpack-deps:bionic

RUN dpkg --add-architecture i386 \
 && apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        libxml-libxml-perl \
        libxml-libxslt-perl \
        python3-sphinx \
        wine32 \
        winetricks \
        xvfb \
 && rm -rf /var/lib/apt/lists/* \
 && useradd --uid 1001 --create-home --shell /bin/bash buildmaster \
 && mkdir -p /opt/cmake \
 && cd /opt \
 && curl -LSo cmake.tar.gz https://cmake.org/files/v3.11/cmake-3.11.1-Linux-x86_64.tar.gz \
 && echo "3aa8d3b53903563bdc13dfec80716c286b2228db36ef65c23bc616f9bb930367  cmake.tar.gz" | sha256sum -c \
 && tar xzCf /opt/cmake /opt/cmake.tar.gz --strip-components=1 \
 && sed -e 's#/Zi#/Z7#g' -i /opt/cmake/share/cmake-3.11/Modules/Platform/Windows-MSVC.cmake \
 && rm -f /opt/cmake.tar.gz

USER buildmaster
WORKDIR /home/buildmaster

ENV PATH=/opt/cmake/bin:$PATH \
    WINEARCH=win32

# Get rid of the "install mono?" and "install gecko?" prompts.
RUN WINEDLLOVERRIDES="mscoree,mshtml=" wine wineboot && wineserver -w

# Install dependencies.
RUN Xvfb :0 -screen 0 1024x768x24 \
  & LOGNAME=buildmaster DISPLAY=:0 winetricks --unattended \
        dotnet452 \
        vcrun2015 \
 && rm -rf /home/buildmaster/.cache/winetricks

# Add native build directory.
RUN mkdir /home/buildmaster/dfhack-native \
 && cd /home/buildmaster/dfhack-native \
 && git clone --depth=1 --recursive https://github.com/DFHack/dfhack.git \
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

# Download clcache.
RUN curl -LSo clcache.zip https://github.com/frerich/clcache/releases/download/v4.1.0/clcache-4.1.0.zip \
 && echo "3d41d98000e1edb705f2237204b943edd95afeb95a6f33c9ff87aa986471cce7  clcache.zip" | sha256sum -c \
 && unzip clcache.zip -d /home/buildmaster/.wine/drive_c/clcache \
 && rm -f clcache.zip

# Set up wine_vcpp
COPY --from=wine_vcpp /home/buildmaster/wine_vcpp /home/buildmaster/wine_vcpp
RUN ln -s /home/buildmaster/wine_vcpp /home/buildmaster/.wine/drive_c/VC

# Add DFHack scripts.
USER root
ADD dfhack-configure dfhack-make wrapper /usr/local/bin/
RUN ln -s wrapper /usr/local/bin/clcache \
 && ln -s wrapper /usr/local/bin/link \
 && ln -s wrapper /usr/local/bin/mt \
 && ln -s wrapper /usr/local/bin/rc
USER buildmaster
ADD toolchain.cmake vcvars_32.bash vcvars_64.bash /home/buildmaster/

ENV WINEDEBUG=-all \
    WINEPATH=C:\\clcache

# Kill mscorsvw so it doesn't cause builds to hang.
#RUN wine C:\\Windows\\Microsoft.NET\\Framework\\v4.0.30319\\ngen.exe executequeueditems
ENV WINEDLLOVERRIDES='ngen.exe,mscorsvw.exe=b'
