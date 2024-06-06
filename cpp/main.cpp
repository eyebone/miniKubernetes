#include "httplib.h"

#include <iostream>
#include <sstream>
#include <memory>
#include <vector>
#include <string>
std::vector<std::string> runKubelet_node(const std::string &req) {
    std::vector<std::string> output;

    // Open the command for reading
    std::shared_ptr<FILE> pipe(popen(req.c_str(), "r"), pclose);
    if (!pipe) {
        std::cerr << "popen() failed!" << std::endl;
        return output;
    }
    // Read the output a line at a time - output it
    char buffer[128];
    while (fgets(buffer, sizeof(buffer), pipe.get()) != nullptr) {
        output.push_back(buffer);
    }
    return output;
}
void handle_apiserver(const httplib::Request &req, httplib::Response &res) {
    std::string data = req.body;
	std::vector<std::string> response = runKubelet_node(data);
	std::string result_str = "";
	for(int i = 0; i < response.size(); i++)
	    result_str += response[i];
	res.set_content(result_str, "text/plain");
}

int main() {
    httplib::Server svr;
    // 接受 api-server 请求
    // :8080/api
    svr.Post("/api", handle_apiserver);
    svr.listen("0.0.0.0", 9099);
}