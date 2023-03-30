package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

var RTM_req = []byte{0x7e, 0x11, 0xf0, 0x0e, 0x00, 0x52, 0x37, 0x1E, 0x00, 0x0f, 0x00, 0x00, 0x00, 0x2b, 0x75, 0x7e}

func errfunc(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func Read_response(file *os.File) (response [76]byte) {
	ch := make(chan [76]byte)
	go func() {
		var arr [76]byte
		for {
			err := binary.Read(file, binary.BigEndian, &arr)
			if err != nil {
				continue
			}
			break
		}
		ch <- arr
	}()

	select {
	case response := <-ch:
		fmt.Println(response)
	case <-time.After(time.Second * 5):
		fmt.Println("Timeout")
	}
	return response
}

func main() {
	// открытие порта
	err := exec.Command("mode", "com2:115200,N,8,1,P").Run()
	errfunc(err)
	file, err := os.OpenFile("COM2", os.O_RDWR, 0700)
	errfunc(err)
	defer file.Close()
	// Отправка запроса
	file.Write(RTM_req)
	// Чтение ответа
	Read_response(file)
}
