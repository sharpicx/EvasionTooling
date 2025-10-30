#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import socket
import sys
import threading
from datetime import datetime

open_ports = []
lock = threading.Lock()

def scan_port(target, port):
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        socket.setdefaulttimeout(1)
        
        result = s.connect_ex((target, port))
        
        if result == 0:
            with lock:
                print("Port {}: Open".format(port))
                open_ports.append(port)
            
        s.close()
    except:
        pass

def scan_target(target, start_port=1, end_port=65535, threads=100):
    print("-" * 60)
    print("Scanning target: {}".format(target))
    print("Scanning ports: {}-{}".format(start_port, end_port))
    print("Scanning started at: {}".format(datetime.now()))
    print("-" * 60)
    
    try:
        target_ip = socket.gethostbyname(target)
    except socket.gaierror:
        print("Hostname could not be resolved. Exiting")
        sys.exit()
    
    # 创建端口范围
    ports = range(start_port, end_port + 1)
    
    # 创建线程池
    thread_list = []
    
    # 分配端口给线程
    for i in range(0, len(ports), threads):
        batch = ports[i:i+threads]
        
        for port in batch:
            thread = threading.Thread(target=scan_port, args=(target_ip, port))
            thread_list.append(thread)
            thread.start()
        
        # 等待当前批次的所有线程完成
        for thread in thread_list:
            thread.join()
        
        thread_list = []
        
        # 显示进度
        progress = float(i + len(batch)) / len(ports) * 100
        sys.stdout.write("\rProgress: {:.2f}%".format(progress))
        sys.stdout.flush()
    
    print("\n" + "-" * 60)
    print("Scanning completed at: {}".format(datetime.now()))
    print("Open ports: {}".format(sorted(open_ports)))
    print("-" * 60)

def main():
    # 检查命令行参数
    if len(sys.argv) < 2:
        print("Usage: python3 port_scanner.py <target> [start_port] [end_port] [threads]")
        print("Example: python3 port_scanner.py example.com 1 65535 100")
        sys.exit()
    
    target = sys.argv[1]
    
    # 设置默认值
    start_port = 1
    end_port = 65535
    threads = 100
    
    # 解析可选参数
    if len(sys.argv) >= 3:
        start_port = int(sys.argv[2])
    if len(sys.argv) >= 4:
        end_port = int(sys.argv[3])
    if len(sys.argv) >= 5:
        threads = int(sys.argv[4])
    
    # 验证端口范围
    if start_port < 1 or end_port > 65535 or start_port > end_port:
        print("Invalid port range. Please use ports between 1 and 65535.")
        sys.exit()
    
    # 验证线程数
    if threads < 1:
        print("Thread count must be at least 1.")
        sys.exit()
    
    scan_target(target, start_port, end_port, threads)

if __name__ == "__main__":
    main()
