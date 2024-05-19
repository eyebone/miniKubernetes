#pragma once

#include <iostream>
#include <unordered_map>
#include <vector>
#include <filesystem>
#include <fstream>
#include <sstream>
#include <stdlib.h>

class PodSpec {
private:
    // volume list

    // NodeName
    std::string node_name;

    struct Container{
        std::string name;
        std::string image;
        // 其他容器属性
    };


    // container list
    std::unordered_map<std::string, Container> containers;
public:

    PodSpec() {
        node_name = "";
        if (!containers.empty())
            containers.clear();
    }
};

class PodStatus {
private:
    // Host ip, if not , it`s nullptr;
    std::string HostIp;
    // description of the Pod,
    // Pending, Running, Succeeded, Failed, Unknown
    std::string Phase;
    // Pod Ip, if not, it`s nullptr;
    std::string PodIp;

public:
    PodStatus() {
        HostIp = nullptr;
        Phase = "Unknown";
        PodIp = nullptr;
    }
};

/*
 * 元数据结构体
 * */
struct Metadata {
    std::string name;
    std::string _namespace;
    std::string uid;

    Metadata() {}

    Metadata(std::string n, std::string s, std::string u) : name(std::move(n)), _namespace(std::move(s)),
                                                            uid(std::move(u)) {}
};

class Pod {
private:
    std::string kind;
    std::string api_version;
    Metadata metadata;
    PodSpec pod_spec;
    PodStatus pod_status;

    //
    PodSpec podSpec;
public:
    Pod() {
        kind = "Pod";

    }

    Pod(const std::string &file_name) {
        std::string path = "../config/pod/" + file_name;
        if (!std::filesystem::exists(path)) {
            std::cerr << "[ERROR]";
            std::cout << "   the path to the yaml file of pod is error or the file is not exists.\n";
            return;
        }
        if (!PodCreate(path)) {
            std::cerr << "[ERROR]";
            std::cout << "   Pod Create ERROR.\n";
            return;
        }
    }

private:

    /**
     * 我们可以看到 Pod 的创建过程其实是比较简单的，
     * 首先计算 Pod 规格和沙箱的变更，然后停止
     * 可能影响这一次创建或者更新的容器，
     * 最后依次创建沙盒、初始化容器和常规容器。
     * */
    // 计算 Pod 中沙盒和容器的变更
    bool CalculateSandboxContainer();

    // 强制停止 Pod 对应的沙盒
    bool StopSandboxByForce();

    // 强制停止所有不应该运行的容器
    bool StopContainerByForce();

    // 为 Pod 创建新的沙盒
    bool CreateSandbox();

    /**
     * pod 启动时运行，初始化配置
     * */
    bool initContainer(const std::string &file_name);

    //检查容器是否存在
    bool isContainerExists(std::string name);

public:
    /**
     * 该函数用于创建pod
     * */
    bool PodCreate(const std::string file_name) {
        return (CalculateSandboxContainer() && StopSandboxByForce() && StopContainerByForce() && CreateSandbox() &&
                initContainer(file_name));
    }


};

