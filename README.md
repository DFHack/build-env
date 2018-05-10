# DFHack Build Environment

*Currently, this does not support Windows builds. I'm trying to fix that. You can see my progress here: <https://what.thedailywtf.com/post/1337023>*

*ccache is disabled for Mac OS X cross compilation until I can figure out what's causing it to break.*

## Contents

### GCC 4.8 image

- Ubuntu Trusty (14.04 LTS)
- buildpack-deps (see [Docker Hub](https://hub.docker.com/_/buildpack-deps/) description for details)
- GNU C/C++ compilers (gcc and g++)
  - Version 4.8.5
  - 32-bit and 64-bit
  - Linux and Mac OS X
  - Minimum OS X version 10.6
- ccache (intended to have cache directory stored outside the container)
- CMake version 3.11 or later
- Google protocol buffer compiler (shim DFHack native build directory at `/home/buildmaster/dfhack-native`)
- Perl with `XML::LibXML` and `XML::LibXSLT` (required for df-structures)
- OpenGL headers and libraries (required for Stonesense)
- Sphinx (used to build DFHack documentation)
- zlib for 32-bit and 64-bit Linux and Mac OS X
- libSDL.so for 32-bit and 64-bit Linux

### Latest image

- Ubuntu Bionic (18.04 LTS)
- buildpack-deps (see [Docker Hub](https://hub.docker.com/_/buildpack-deps/) description for details)
- GNU C/C++ compilers (gcc and g++)
  - Version 7.3.0 or later
  - 32-bit and 64-bit
  - Linux and Mac OS X
  - Minimum OS X version 10.6
- ccache (intended to have cache directory stored outside the container)
- CMake version 3.11 or later
- Google protocol buffer compiler (shim DFHack native build directory at `/home/buildmaster/dfhack-native`)
- Perl with `XML::LibXML` and `XML::LibXSLT` (required for df-structures)
- OpenGL headers and libraries (required for Stonesense)
- Sphinx (used to build DFHack documentation)
- zlib for 32-bit and 64-bit Linux and Mac OS X
- libSDL.so for 32-bit and 64-bit Linux

## Special Thanks

- Mac OS X cross-compiler: <https://github.com/tpoechtrager/osxcross>
- Mac OS X SDK mirror: <https://github.com/phracker/MacOSX-SDKs>
