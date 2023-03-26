package main

import (
	"fmt"

	"github.com/goburrow/modbus"
)

func main() {
	var comport string
	var Slave, temp byte
	var FirstReg uint16
	fmt.Print("Введите Com-порт:")
	fmt.Scan(&comport)
	fmt.Print("Введите modbus адрес ПЛК:")
	fmt.Scan(&Slave)
	fmt.Print("Введите первый регистр:")
	fmt.Scan(&FirstReg)

	handler := modbus.NewRTUClientHandler("COM" + comport)
	handler.BaudRate = 115200
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = Slave
	handler.Timeout = 1000

	err := handler.Connect()
	defer handler.Close()
	client := modbus.NewClient(handler)
	mb_arr, err := client.ReadHoldingRegisters(FirstReg, 32) //412
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
