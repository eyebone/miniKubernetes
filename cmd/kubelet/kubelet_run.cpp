#include <iostream>
#include <cstdlib>
#include <fstream>

#include "../../entity/pod.h"

using namespace std;
int main(int argc, char **argv) {
    if(argc == 1) {
        system("cat ../../doc/help/help_kubelet.txt");
        return 0;
    }
    for(int i = 1; i < argc; ++i) {
        if(argv[i] == "-f") return 0;
        else if(argv[i] == "-v") return 0;
    }

    return 0;
}