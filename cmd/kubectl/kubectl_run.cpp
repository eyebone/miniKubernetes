#include <iostream>
#include <cstdlib>
#include <fstream>
#include <string>
#include <stdlib.h>

//#include "../../entity/pod.h"

using namespace std;
int main(int argc, char **argv) {

    if(argc < 4) {
        cerr << "[ERROR]\n";
        cout << "   The token numbers of argvs is not enough.\n";
        system("cat ../../doc/help/help_kubectl.txt");
        return 0;
    }
    // kubectl [command] [TYPE] [NAME] [flags]
    std::string COMMAND = argv[1];
    std::string TYPE = argv[2];
    std::string NAME = argv[3];
    std::string command_to_API = "curl -X POST http://localhost:6033 " +
            COMMAND + '-' + TYPE + '-' + NAME + '\n';

    system(command_to_API.c_str());

    int flags = argc - 4;


    return 0;
}