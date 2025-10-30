package main

import (
	"io/ioutil"
	"net/http"
	"fmt"
	"syscall"
	"unsafe"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"runtime"
	"sync/atomic"
	"os"
)

import "C"

// internal maps
var (
	listeners   = make(map[int]net.Listener)
	listenMutex sync.Mutex
	nextHandle  = 1
)

func hostPort(h string, p int) string { return net.JoinHostPort(h, strconv.Itoa(p)) }

//export StartTCPListener
func StartTCPListener(host *C.char, port C.int, verbose C.int) C.int {
	return startListener(host, port, verbose, false)
}

//export StartOneShotTCPListener
func StartOneShotTCPListener(host *C.char, port C.int, verbose C.int) C.int {
	return startListener(host, port, verbose, true)
}

func startListener(host *C.char, port C.int, verbose C.int, oneShot bool) C.int {
	goHost := C.GoString(host)
	goPort := int(port)
	bind := hostPort(goHost, goPort)

	ln, err := net.Listen("tcp", bind)
	if err != nil {
		if verbose != 0 {
			fmt.Printf("[ERR] listen %s failed: %v\n", bind, err)
		}
		return C.int(-1)
	}

	listenMutex.Lock()
	h := nextHandle
	nextHandle++
	listeners[h] = ln
	listenMutex.Unlock()

	if verbose != 0 {
		mode := "persistent"
		if oneShot {
			mode = "one-shot"
		}
		fmt.Printf("[INFO] listening on %s (handle=%d, mode=%s)\n", bind, h, mode)
	}

	go func(h int, ln net.Listener, vb bool, oneShot bool) {
		defer func() {
			listenMutex.Lock()
			if _, ok := listeners[h]; ok {
				delete(listeners, h)
			}
			listenMutex.Unlock()
		}()

		for {
			conn, err := ln.Accept()
			if err != nil {
				if vb {
					fmt.Printf("[ERR] accept (handle=%d): %v\n", h, err)
				}
				ln.Close()
				return
			}
			if vb {
				fmt.Printf("[INFO] accepted %s -> %s (handle=%d)\n", conn.RemoteAddr(), conn.LocalAddr(), h)
			}
			go func(c net.Conn, vb bool) {
				defer c.Close()
				n, err := io.Copy(c, c)
				if vb {
					if err != nil {
						fmt.Printf("[ERR] conn %s: %v\n", c.RemoteAddr(), err)
					} else {
						fmt.Printf("[INFO] conn %s closed (echoed %d bytes)\n", c.RemoteAddr(), n)
					}
				}
			}(conn, vb)

			if oneShot {
				if vb {
					fmt.Printf("[INFO] one-shot listener handle=%d closing after first connection\n", h)
				}
				ln.Close()
				listenMutex.Lock()
				if _, ok := listeners[h]; ok {
					delete(listeners, h)
				}
				listenMutex.Unlock()
				return
			}
		}
	}(h, ln, verbose != 0, oneShot)

	return C.int(h)
}

//export StopTCPListener
func StopTCPListener(handle C.int) C.int {
	h := int(handle)
	listenMutex.Lock()
	ln, ok := listeners[h]
	if ok {
		delete(listeners, h)
	}
	listenMutex.Unlock()
	if !ok {
		fmt.Printf("[WARN] invalid handle: %d\n", h)
		return C.int(-1)
	}
	err := ln.Close()
	if err != nil {
		fmt.Printf("[ERR] close listener %d: %v\n", h, err)
		return C.int(-1)
	}
	fmt.Printf("[INFO] listener %d stopped\n", h)
	return C.int(0)
}

//export SendTCP
func SendTCP(host *C.char, port C.int, data *C.char, timeoutMs C.int, verbose C.int) C.int {
	goHost := C.GoString(host)
	goPort := int(port)
	payload := C.GoString(data)
	addr := hostPort(goHost, goPort)

	if verbose != 0 {
		fmt.Printf("[INFO] connecting to %s ...\n", addr)
	}

	d := net.Dialer{}
	if timeoutMs > 0 {
		d.Timeout = time.Duration(int(timeoutMs)) * time.Millisecond
	}
	conn, err := d.Dial("tcp", addr)
	if err != nil {
		if verbose != 0 {
			fmt.Printf("[ERR] connect failed: %v\n", err)
		}
		return C.int(-1)
	}
	defer conn.Close()

	if verbose != 0 {
		fmt.Printf("[INFO] connected to %s\n", addr)
	}

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	n, err := io.Copy(conn, strings.NewReader(payload))
	if err != nil {
		if verbose != 0 {
			fmt.Printf("[ERR] send failed: %v\n", err)
		}
		return C.int(-1)
	}
	if verbose != 0 {
		fmt.Printf("[INFO] sent %d bytes\n", n)
	}

	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	nr, err := conn.Read(buf)
	if err != nil {
		if verbose != 0 {
			fmt.Printf("[WARN] read after send: %v\n", err)
		}
		return C.int(0)
	}
	if verbose != 0 && nr > 0 {
		fmt.Printf("[INFO] recv: %q\n", string(buf[:nr]))
	}
	return C.int(0)
}

// Convenience exported names for P/Invoke consistency
//export NetcatStartTCPListener
func NetcatStartTCPListener(host *C.char, port C.int, verbose C.int) C.int {
	return StartTCPListener(host, port, verbose)
}

//export NetcatStartOneShotTCPListener
func NetcatStartOneShotTCPListener(host *C.char, port C.int, verbose C.int) C.int {
	return StartOneShotTCPListener(host, port, verbose)
}

//export NetcatStopTCPListener
func NetcatStopTCPListener(handle C.int) C.int {
	return StopTCPListener(handle)
}

//export NetcatSendTCP
func NetcatSendTCP(host *C.char, port C.int, data *C.char, timeoutMs C.int, verbose C.int) C.int {
	return SendTCP(host, port, data, timeoutMs, verbose)
}

//export PortScan
func PortScan(host *C.char, start C.int, end C.int, timeoutMs C.int, verbose C.int, logPath *C.char) C.int {
	h := C.GoString(host)
	s := int(start)
	e := int(end)
	if s < 1 {
		s = 1
	}
	if e > 65535 {
		e = 65535
	}
	if s > e {
		s, e = e, s
	}
	to := time.Duration(int(timeoutMs)) * time.Millisecond
	cpu := runtime.NumCPU()
	concurrency := cpu * 200
	if concurrency < 64 {
		concurrency = 64
	}
	if concurrency > 4096 {
		concurrency = 4096
	}
	var outWriter io.Writer = os.Stdout
	var logFile *os.File
	if lp := C.GoString(logPath); lp != "" {
		f, err := os.OpenFile(lp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err == nil {
			logFile = f
			outWriter = f
		} else {
			fmt.Fprintf(os.Stderr, "[WARN] cannot open log file %s: %v\n", lp, err)
		}
	}
	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()
	var outMutex sync.Mutex
	logf := func(format string, a ...interface{}) {
		outMutex.Lock()
		fmt.Fprintf(outWriter, format, a...)
		outMutex.Unlock()
	}
	if verbose != 0 {
		logf("[INFO] portscan %s ports %d-%d (timeout=%s, concurrency=%d)\n", h, s, e, to, concurrency)
	}
	var openCount int32
	var scanned int32
	sem := make(chan struct{}, concurrency)
	wg := sync.WaitGroup{}
	dialer := &net.Dialer{Timeout: to}
	results := make(chan int, 256)
	for port := s; port <= e; port++ {
		sem <- struct{}{}
		wg.Add(1)
		go func(p int) {
			defer func() {
				<-sem
				wg.Done()
				atomic.AddInt32(&scanned, 1)
			}()
			addr := hostPort(h, p)
			conn, err := dialer.Dial("tcp", addr)
			if err == nil {
				conn.Close()
				atomic.AddInt32(&openCount, 1)
				results <- p
				if verbose != 0 {
					logf("[OPEN] %s:%d\n", h, p)
				}
				return
			}
			if verbose != 0 {
				if strings.Contains(err.Error(), "refused") || strings.Contains(err.Error(), "i/o timeout") {
				} else {
					logf("[DBG] %s:%d - %v\n", h, p, err)
				}
			}
		}(port)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	for range results {
	}
	if verbose != 0 {
		logf("[INFO] scan complete: scanned=%d open=%d\n", atomic.LoadInt32(&scanned), atomic.LoadInt32(&openCount))
	}
	return C.int(openCount)
}

//export Tembak
func Tembak(url *C.char) {
    goURL := C.GoString(url)

    resp, err := http.Get(goURL)
    if err != nil {
        return
    }
    defer resp.Body.Close()

    _, _ = ioutil.ReadAll(resp.Body)
}

//export HelloWorld
func HelloWorld() *C.char {
    message := "Hello from Go DLL!"
    return C.CString(message)
}

//export ShowMessage
func ShowMessage(title, message *C.char) {
    titleStr := C.GoString(title)
    messageStr := C.GoString(message)

    user32, err := syscall.LoadLibrary("user32.dll")
    if err != nil {
        fmt.Printf("Error loading user32.dll: %v\n", err)
        return
    }
    defer syscall.FreeLibrary(user32)

    messageBoxProc, err := syscall.GetProcAddress(user32, "MessageBoxW")
    if err != nil {
        fmt.Printf("Error finding MessageBoxW: %v\n", err)
        return
    }

    titlePtr, err := syscall.UTF16PtrFromString(titleStr)
    if err != nil {
        fmt.Printf("Invalid title string: %v\n", err)
        return
    }

    messagePtr, err := syscall.UTF16PtrFromString(messageStr)
    if err != nil {
        fmt.Printf("Invalid message string: %v\n", err)
        return
    }

    ret, _, callErr := syscall.Syscall6(
        uintptr(messageBoxProc),
        4,
        0,
        uintptr(unsafe.Pointer(messagePtr)),
        uintptr(unsafe.Pointer(titlePtr)),
        0,
        0,
        0,
    )
    if callErr != syscall.Errno(0) {
        fmt.Printf("Error calling MessageBoxW: %v\n", callErr)
        return
    }
    _ = ret
}

func main() {
}
