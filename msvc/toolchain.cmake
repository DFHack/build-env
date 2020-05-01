set(CMAKE_SYSTEM_NAME Windows)
set(BUILD_SHARED_LIBS ON)
set(LIBTYPE SHARED)
set(CMAKE_CXX_COMPILER_ID MSVC)
set(CMAKE_CXX_PLATFORM_ID Windows)
set(CMAKE_CXX_COMPILER /usr/local/bin/clcache)
set(CMAKE_C_COMPILER_ID MSVC)
set(CMAKE_C_PLATFORM_ID Windows)
set(CMAKE_C_COMPILER /usr/local/bin/clcache)
set(CMAKE_LINKER /usr/local/bin/link)
set(CMAKE_BUILD_TYPE Release CACHE STRING "Release") #|RelWithDebInfo
set(CMAKE_CROSS_COMPILING ON)
set(MSVC_VERSION 1900)
set(MSVC10 ON) # for dfplex
