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
FROM buildpack-deps:bionic

RUN dpkg --add-architecture i386 \
 && apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        wine-development \
        wine32-development \
        wine64-development \
        winetricks \
        xvfb \
 && rm -rf /var/lib/apt/lists/* \
 && useradd --uid 1001 --create-home --shell /bin/bash buildmaster

USER buildmaster
WORKDIR /home/buildmaster

RUN LOGNAME=buildmaster WINEDEBUG=-all xvfb-run -a winetricks --unattended \
        dotnet462 \
        vcrun2017 \
	win10 \
 && rm -rf /home/buildmaster/.cache/winetricks

#COPY --from=msvc-downloader /download-msvc.exe /home/buildmaster/download-msvc.exe
#
#RUN wine download-msvc.exe

RUN wget -O vs_BuildTools.exe https://aka.ms/vs/15/release/vs_BuildTools.exe

RUN wine vs_BuildTools.exe --passive --wait --norestart --nocache \
	--installPath C:\\BuildTools \
        --add Microsoft.VisualStudio.Workload.VCTools \
        --add Microsoft.VisualStudio.Component.VC.140 \
        --add Microsoft.VisualStudio.Component.WinXP

RUN curl -LSo clcache.zip https://github.com/frerich/clcache/releases/download/v4.1.0/clcache-4.1.0.zip \
 && echo "3d41d98000e1edb705f2237204b943edd95afeb95a6f33c9ff87aa986471cce7  clcache.zip" | sha256sum -c \
 && unzip clcache.zip -d /home/buildmaster/.wine/drive_c/clcache \
 && rm -f clcache.zip

ADD dfhack-configure dfhack-make /usr/local/bin/
