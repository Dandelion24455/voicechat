#include <portaudio.h>
#include <opus/opus.h>
#include <vector>
#include <mutex>

class AudioProcessor {
public:
    AudioProcessor();
    ~AudioProcessor();
    
    // 初始化音频设备
    bool init(int sample_rate = 48000, int frames_per_buffer = 960);
    
    // 音频回调函数
    static int audioCallback(
        const void* input, void* output,
        unsigned long frameCount,
        const PaStreamCallbackTimeInfo* timeInfo,
        PaStreamCallbackFlags statusFlags,
        void* userData
    );

    // 编码/解码方法
    std::vector<uint8_t> encode(const short* pcm_data, int frame_size);
    std::vector<short> decode(const uint8_t* opus_data, int data_length);

private:
    PaStream* stream_;
    OpusEncoder* encoder_;
    OpusDecoder* decoder_;
    std::mutex audio_mutex_;
};