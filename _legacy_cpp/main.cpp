#include "audio_processor.h"
#include "udp_client.h"
#include <thread>
#include <atomic>

// 全局标志位控制音频线程
std::atomic<bool> is_running(true);

// 音频处理线程
void audioThread(AudioProcessor& audio, UDPClient& client) {
    const int buffer_size = 960; // 20ms 数据块
    std::vector<short> pcm_buffer(buffer_size);

    while (is_running) {
        // 模拟音频采集（实际需用PortAudio回调）
        // 这里简化为生成静音数据
        std::fill(pcm_buffer.begin(), pcm_buffer.end(), 0);
        
        // 编码并发送
        auto encoded = audio.encode(pcm_buffer.data(), buffer_size);
        client.send(encoded);
        
        // 接收并解码播放
        auto received = client.receive();
        if (!received.empty()) {
            auto decoded = audio.decode(received.data(), received.size());
            // 播放解码后的PCM数据（实际需用PortAudio输出）
        }
        
        std::this_thread::sleep_for(std::chrono::milliseconds(20));
    }
}

int main() {
    try {
        // 初始化音频
        AudioProcessor audio;
        if (!audio.init()) return 1;
        
        // 初始化网络（示例IP和端口）
        boost::asio::io_context io_context;
        UDPClient client(io_context, "127.0.0.1", 12345);
        
        // 启动音频线程
        std::thread audio_thread(audioThread, std::ref(audio), std::ref(client));
        
        // 主线程等待退出
        std::cout << "Press Enter to exit...\n";
        std::cin.get();
        is_running = false;
        audio_thread.join();
        
    } catch (std::exception& e) {
        std::cerr << "Exception: " << e.what() << "\n";
    }
    return 0;
}