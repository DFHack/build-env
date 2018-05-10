FROM buildpack-deps:bionic

RUN curl -sSL https://repos.wine-staging.com/wine/Release.key | apt-key add - \
 && dpkg --add-architecture i386 \
 && apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        software-properties-common \
 && apt-add-repository 'deb https://dl.winehq.org/wine-builds/ubuntu/ artful main' \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
	cabextract \
        winehq-staging \
        xvfb \
        zip \
 && rm -rf /var/lib/apt/lists/* \
 && useradd --uid 1001 --create-home --shell /bin/bash buildmaster \
 && curl -sSL https://raw.githubusercontent.com/Winetricks/winetricks/master/src/winetricks -o /usr/local/bin/winetricks \
 && chmod +x /usr/local/bin/winetricks

USER buildmaster
WORKDIR /home/buildmaster

ENV WINEARCH=win32

RUN WINEDLLOVERRIDES="mscoree,mshtml=" wine wineboot && wineserver -w

RUN LOGNAME=buildmaster WINEDEBUG=-all xvfb-run -a winetricks --unattended \
        dotnet45 \
        vcrun2013 \
	win10 \
 && rm -rf /home/buildmaster/.cache/winetricks

ADD en_visual_studio_community_2015_with_update_3_x86_x64_web_installer_8922963.exe /home/buildmaster/
RUN WINEDEBUG=+all,trace-all wine en_visual_studio_community_2015_with_update_3_x86_x64_web_installer_8922963.exe /CustomInstallPath C:\BuildTools /InstallSelectableItems "NativeLanguageSupport_VCV1;NativeLanguageSupport_XPV1;GitForWindowsx86V9" /Quiet /NoRestart

RUN curl -LSo clcache.zip https://github.com/frerich/clcache/releases/download/v4.1.0/clcache-4.1.0.zip \
 && echo "3d41d98000e1edb705f2237204b943edd95afeb95a6f33c9ff87aa986471cce7  clcache.zip" | sha256sum -c \
 && unzip clcache.zip -d /home/buildmaster/.wine/drive_c/clcache \
 && rm -f clcache.zip

ADD dfhack-configure dfhack-make /usr/local/bin/
