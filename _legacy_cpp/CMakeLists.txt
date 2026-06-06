cmake_minimum_required(VERSION 3.10)
project(VoiceChat)

set(CMAKE_CXX_STANDARD 17)

# ========== 核心配置 ==========
# Homebrew 默认安装路径
set(HOMEBREW_PREFIX "/usr/local")

# 包含所有头文件路径
include_directories(
    ${HOMEBREW_PREFIX}/include
    ${HOMEBREW_PREFIX}/include/opus
)

# 链接所有库路径
link_directories(
    ${HOMEBREW_PREFIX}/lib
)

# ========== 手动查找库文件 ==========
# PortAudio
find_library(PORTAUDIO_LIB portaudio)
if(NOT PORTAUDIO_LIB)
    message(FATAL_ERROR "PortAudio library not found!")
endif()

# Opus
find_library(OPUS_LIB opus)
if(NOT OPUS_LIB)
    message(FATAL_ERROR "Opus library not found!")
endif()

# Boost (需指定具体组件)
find_library(BOOST_SYSTEM_LIB boost_system)
if(NOT BOOST_SYSTEM_LIB)
    message(FATAL_ERROR "Boost.System library not found!")
endif()

# ========== 生成可执行文件 ==========
add_executable(voice_chat
    main.cpp
    audio_processor.cpp
    udp_client.cpp
)

# ========== 链接库 ==========
target_link_libraries(voice_chat
    ${PORTAUDIO_LIB}
    ${OPUS_LIB}
    ${BOOST_SYSTEM_LIB}
    # macOS 必要框架
    -framework CoreAudio
    -framework CoreFoundation
    -framework AudioToolbox
)