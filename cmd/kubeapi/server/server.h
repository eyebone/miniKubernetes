//
// Created by 陆逸凡 on 2024/5/12.
//
#pragma once
#ifndef MINIK8S_SERVER_H
#define MINIK8S_SERVER_H

#include <iostream>
#include <netinet/in.h>
#include <sys/socket.h>

using std::cout;
// localhost:3306
void run_server() {
    cout << "Open K8s Api Server succeed.\n";
    const int BUFFER_SIZE = 102400;
    int slisten = socket(AF_INET, SOCK_STREAM, 0);

    sockaddr_in sin;

    sin.sin_family = AF_INET;
    sin.sin_port = htons(6033);
//        sin.sin_addr.s_addr = htonl(INADDR_ANY);
    sin.sin_addr.s_addr = htonl(INADDR_ANY);
    if (bind(slisten, (struct sockaddr *) &sin, sizeof(sin)) == -1) {
        printf("    bind error!\n");
        fflush(stdout);

    }
    else printf("bind success!\n");
    fflush(stdout);
    if (listen(slisten, 5) == -1) {
        printf("    listen error!\n");
        fflush(stdout);
        exit(0);
    }
    else printf("listen success!\n");
    fflush(stdout);
    int sclient;
    sockaddr_in client_add;
    socklen_t naddrlen = sizeof(client_add);
    char revdata[BUFFER_SIZE], command[BUFFER_SIZE], send_result[BUFFER_SIZE];

    while (1) {
        printf("start run.\n");
        fflush(stdout);
        memset(command, 0, sizeof(command));
        memset(revdata, 0, sizeof(revdata));
        sclient = accept(slisten, (struct sockaddr *) &client_add, &naddrlen);
        printf("socket accept success");
        fflush(stdout);
        if (sclient == -1) {
            printf("Socket:Accept Error!");
            fflush(stdout);
            continue;
        }
        printf("socket accept success");
        fflush(stdout);
        recv(sclient, revdata, BUFFER_SIZE, 0);
        std::cout << "recieve command: " << revdata << ", length: " << strlen(revdata) << std::endl;


        //revdata is received
        std::string res = revdata;
        std::cout << "return: " << res << std::endl;
        fflush(stdout);
        //res is sent
        send(sclient, res.c_str(), res.length(), 0);
    }
}

#endif //MINIK8S_SERVER_H
