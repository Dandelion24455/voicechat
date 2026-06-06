#include <boost/asio.hpp>
#include <vector>

class UDPClient {
public:
    UDPClient(boost::asio::io_context& io_context, 
             const std::string& host, int port);
    
    void send(const std::vector<uint8_t>& data);
    std::vector<uint8_t> receive();
    
private:
    boost::asio::ip::udp::socket socket_;
    boost::asio::ip::udp::endpoint receiver_endpoint_;
};