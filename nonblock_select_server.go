package main

import (
	"fmt"
	"log"
	"net"
	"syscall"
	"time"
)

func FD_CLR(fd int, set *syscall.FdSet) {
	index := fd / 64
	offset := fd % 64
	set.Bits[index] &^= (1 << offset)
}

func FD_ZERO(set *syscall.FdSet) {
	for i := range set.Bits {
		set.Bits[i] = 0
	}
}

func FD_SET(fd int, set *syscall.FdSet) {
	set.Bits[fd/64] |= 1 << (uint(fd) % 64)
}

func FD_ISSET(fd int, set *syscall.FdSet) bool {
	return set.Bits[fd/64]&(1<<(uint(fd)%64)) != 0
}

func connHandler(conn net.Conn) {

	var readcnt int = 0

	for {

		fd := getFd(conn)
		n := IsSetReadData(fd)

		buffer := make([]byte, 1024)
		if n > 0 {
			cnt, err := conn.Read(buffer)

			readcnt++
			if readcnt > 10000 {
				readcnt = 0
			}
			if err != nil {
				log.Println("err.. ", err)
			} else {
				log.Println("read:", buffer[:cnt])
			}

			if readcnt%10 == 0 {
				// echo (10번중 1번만)
				conn.Write(buffer[:cnt])
			}

		} else if n == 0 {
			log.Println("time out..")
		}

		time.Sleep(time.Millisecond * 100)
	}
}

func main() {
	// 서버 소켓 생성

	servSock, err := net.Listen("tcp", ":11001")
	if err != nil {
		fmt.Println("Failed to Listen : ", err)
	}
	defer servSock.Close()

	for {
		conn, err := servSock.Accept()
		if err != nil {
			fmt.Println("Failed to Accept : ", err)
			continue
		}
		log.Println("Client connected..")

		go connHandler(conn)
	}

}

func IsSetReadData(fd int) int {
	var readFds syscall.FdSet
	var expFds syscall.FdSet

	FD_ZERO(&readFds)
	FD_ZERO(&expFds)
	FD_SET(fd, &readFds)
	FD_SET(fd, &expFds)

	tv := syscall.NsecToTimeval(10 * time.Millisecond.Nanoseconds()) // 타임아웃을 10밀리초로 설정

	n, err := syscall.Select(fd+1, &readFds, nil, &expFds, &tv)

	if n < 0 || err != nil {
		return -1
	}

	if FD_ISSET(fd, &expFds) {
		return -1
	}
	if FD_ISSET(fd, &readFds) {
		return 1
	}
	return 0
}

func getFd(conn net.Conn) int {
	tcpConn := conn.(*net.TCPConn)
	file, err := tcpConn.File()
	if err != nil {
		panic(err)
		//return -1
	}
	return int(file.Fd())
}
