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

var RTM_req_exam = []byte{0x7e, 0x11, 0xf0, 0x0e, 0x00, 0x52, 0x37, 0x1E, 0x00, 0x0f, 0x00, 0x00, 0x00, 0x2b, 0x75, 0x7e}                                                                                                                                                                  //4700
var RTM_res_exam = []byte{126, 17, 240, 74, 0, 82, 55, 30, 144, 92, 18, 0, 0, 17, 0, 182, 227, 0, 100, 0, 0, 0, 5, 12, 0, 0, 0, 0, 0, 0, 0, 1, 1, 246, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 209, 202, 126} //4700

func Read_response(file *os.File) (res [76]byte) { //Чтение ответа
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
		fmt.Print("\n Ответ получен:  \n")
		return res
	case <-time.After(time.Second * 2):
		log.Fatal("Timeout, нет ответа от ПЛК")
	}
	return res
}

func request(sl uint16, arch uint32, file *os.File) { //Формирование и отправка запроса
	request_buff := bytes.Buffer{}
	request_buff.Write([]byte{0x11, 0xf0, 0x0e, 0x00, 0x52, 0x37})
	binary.Write(&request_buff, binary.LittleEndian, sl)
	binary.Write(&request_buff, binary.LittleEndian, arch)
	binary.Write(&request_buff, binary.LittleEndian, CRC(request_buff.Bytes()))
	// Отправка запроса
	file.Write([]byte{0x7e}) //Пилот протокола RTM
	request_buff.WriteTo(file)
	file.Write([]byte{0x7e}) //Пилот протокола RTM
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

func errfunc(err error) { //вывод ошибки в лог при наличии
	if err != nil {
		log.Fatal(err.Error())
	}
}

func makeTable(res []byte, AInd uint8, conf config) {
	var B byte
	var U16 uint16
	var U32 uint32
	var F32 float32

	Archive := conf.Archives[AInd]
	var t, bytecount int //№ байта в архиве, кол-во байт в записи
	//Разбор архива
	for i, stroke := range Archive.Data { // по строкам данных архива в yaml
		fmt.Printf("  %2d |  %02X ", t, res[t]) //вывод байта и его №
		switch Archive.Data[i].Mode {
		case 1: //1 byte
			B = res[t]
			fmt.Printf("|    %10d     | %s \n", B, stroke.Text)
			bytecount = 1
		case 2: //2 byte
			U16 = binary.LittleEndian.Uint16(res[t : t+3])
			fmt.Printf("|    %10d     | %s \n", U16, stroke.Text)
			bytecount = 2
		case 3: //4 byte - REAL
			buf := bytes.NewReader(res[t : t+4])
			binary.Read(buf, binary.LittleEndian, &F32)
			fmt.Printf("| %18F | %s \n", F32, stroke.Text)
			bytecount = 4
		case 4: //4 byte - DWORD
			U32 = binary.LittleEndian.Uint32(res[t : t+4])
			fmt.Printf("|    %10d     | %s \n", U32, stroke.Text)
			bytecount = 4
		case 5: //4 byte - TIME
			U32 = binary.LittleEndian.Uint32(res[t : t+4])
			fmt.Printf("|    %10d     | %s   %s \n", U32, time.Unix(int64(U32), 0).Format("01-02-2006 15:04:05"), stroke.Text)
			bytecount = 3
		}

		for t++; bytecount > 1; bytecount-- { // Вывод пустых строк
			fmt.Printf("  %2d |  %02X | \n", t, res[t])
			t++
		}
	}
}

func main() {

	var ArchType, Archindex uint8
	var needArch uint32

	conf, err := getConfigYAML("config.yml")
	if err != nil {
		log.Fatal(err.Error())
	}

	//Ввод значений

	fmt.Print("Введите Com-порт: ")
	fmt.Scanln(&conf.DefsCom)
	fmt.Print("Введите modbus адрес ПЛК: ")
	fmt.Scanln(&conf.DefsPlc)
	// открытие порта
	exec.Command("mode", "com"+conf.DefsCom+":"+conf.DefsBaud+",N,8,1").Run() //Настройка порта
	file, err := os.OpenFile("COM"+conf.DefsCom, os.O_RDWR, 0700)
	errfunc(err)
	defer file.Close()

	for {
		fmt.Print("Введите № требуемого архива: ")
		if fmt.Scanln(&needArch); needArch == 0 { //Запрос последнего архива.
			needArch = 2147483647
		}

		request(conf.DefsPlc, needArch, file) //Формирование и отправка запроса
		RTMresponse := Read_response(file)    // Чтение ответа

		//RTMresponse := RTM_res_exam
		for _, t := range RTMresponse { // вывод ответа
			fmt.Printf("%02X ", t)
		}
		if RTMresponse[13] == 17 {
			ArchType = uint8(RTMresponse[22])
		} else {
			ArchType = uint8(RTMresponse[13])
		}
		//Поиск архива
		for ind, tt := range conf.Archives {
			if tt.Type == ArchType {
				Archindex = uint8(ind)
				break
			}
		}

		fmt.Println("\n -------------------- ТИП АРХИВА: ", conf.Archives[Archindex].Name, "--------------------\n ")
		fmt.Println("  №  | HEX |        DEC        |  TEXT  ")
		fmt.Println(" ----+-----+-------------------+----------------------------------------------------------")
		makeTable(RTMresponse[9:75], 0, conf) //шапка
		fmt.Println(" ----+-----+-------------------+----------------------------------------------------------")
		makeTable(RTMresponse[20:75], Archindex, conf) //архив
	}

}
