package main

import (
	"encoding/binary"
	"fmt"

	"github.com/goburrow/modbus"
)

const reg_quantity = 32

var mb_arr []byte

/*
	type archive struct {
		ArchNumber uint16
		ArchType   byte
		LastFlag   byte
		ArchTime   uint16
		data       [53]byte
	}
*/
func main() {
	var comport string
	var slave byte
	var firstReg uint16
	var needArch uint32
	fmt.Print("Введите Com-порт:")
	fmt.Scan(&comport)
	fmt.Print("Введите modbus адрес ПЛК:")
	fmt.Scan(&slave)
	fmt.Print("Введите первый регистр:")
	fmt.Scan(&firstReg)
	for {
		fmt.Print("Введите номер требуемого архива:")
		fmt.Scan(&needArch)
		ReadArch(needArch, comport, slave, firstReg)
	}

}

func ReadArch(need uint32, port string, slave byte, fReg uint16) {
	var temp byte
	//	var bt [2]byte = [2]byte{5, 0}
	handler := modbus.NewRTUClientHandler("COM" + port)
	handler.BaudRate = 115200
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = slave
	handler.Timeout = 1000
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, need)

	//defer handler.Close()
	client := modbus.NewClient(handler)
	_, err := client.WriteMultipleRegisters(410, 2, buf) //Write Archive number
	if err != nil {
		fmt.Println(err)
		return
	}
	mb_arr, err = client.ReadHoldingRegisters(fReg+2, reg_quantity) //412 for Cilk
	handler.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	//byte swap, cause modbus response data swaped
	for i := 0; i < 32; i++ {
		temp = mb_arr[i*2]
		mb_arr[i*2] = mb_arr[i*2+1]
		mb_arr[i*2+1] = temp
	}
	fmt.Println(mb_arr)
}
