package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func errfunc(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func Read_response(file *os.File) (res [76]byte) {
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
	case res := <-ch:
		fmt.Println("Ответ получен")
		return res
	case <-time.After(time.Second * 5):
		fmt.Println("Timeout")
	}
	return res
}

func CRC(buff []byte) uint16 { //RTM CRC (CRC-16/XMODEM)
	var crc uint16
	for _, num := range buff {
		crc = crc ^ (uint16(num) << 8)
		for i := 8; i > 0; i-- {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc = crc << 1
			}
		}
	}
	return crc
}

func main() {
	var comport string
	var needArch uint32
	var slave uint16
	fmt.Print("Введите Com-порт: ")
	fmt.Scan(&comport)
	fmt.Print("Введите modbus адрес ПЛК: ")
	fmt.Scan(&slave)
	fmt.Print("Введите № требуемого архива: ")
	fmt.Scan(&needArch)

	// открытие порта
	exec.Command("mode", "com"+comport+":19200,N,8,1").Run() //Настройка порта
	file, err := os.OpenFile("COM"+comport, os.O_RDWR, 0700)
	errfunc(err)
	defer file.Close()
	// Формирование запроса
	request_buff := bytes.Buffer{}
	request_buff.Write([]byte{0x11, 0xf0, 0x0e, 0x00, 0x52, 0x37})
	binary.Write(&request_buff, binary.LittleEndian, slave)
	binary.Write(&request_buff, binary.LittleEndian, needArch)
	binary.Write(&request_buff, binary.LittleEndian, CRC(request_buff.Bytes()))
	// Отправка запроса
	file.Write([]byte{0x7e}) //Пилот протокола RTM
	request_buff.WriteTo(file)
	file.Write([]byte{0x7e}) //Пилот протокола RTM
	// Чтение ответа
	RTMresponse := Read_response(file)
	fmt.Print(RTMresponse)
}
