package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"gopkg.in/yaml.v3"
)

type configfile struct {
	Com        string   `yaml:"defs_com"`
	Baud       string   `yaml:"defs_baud"`
	Plc        uint16   `yaml:"defs_plcaddress"`
	TitleBytes []int    `yaml:"R7title_bytecount"`
	TitleSpec  []string `yaml:"R7title_comment"`
}

var Conf configfile

func yamlRead() { //Чтение файла настроек и архивов
	yfile, err := ioutil.ReadFile("configs.yml")
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Файл конфигурации не найден")
		return
	}
	if err = yaml.Unmarshal(yfile, &Conf); err != nil {
		fmt.Println(err.Error())
		fmt.Println("Ошибка в структуре файла конфигурации")
		return
	}
	//	fmt.Print(R7)
	fmt.Println("Файл настроек прочитан")
}

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
		//Проверка на валидность пакета
		fmt.Println("Ответ получен")
		return res
	case <-time.After(time.Second * 5):
		fmt.Println("Timeout")
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

func main() {
	yamlRead() //Чтение YAML
	//Ввод значений
	var needArch uint32
	fmt.Print("Введите Com-порт: ")
	fmt.Scanln(&Conf.Com)
	fmt.Print("Введите modbus адрес ПЛК: ")
	fmt.Scanln(&Conf.Plc)
	// открытие порта
	exec.Command("mode", "com"+Conf.Com+":"+Conf.Baud+",N,8,1").Run() //Настройка порта
	file, err := os.OpenFile("COM"+Conf.Com, os.O_RDWR, 0700)
	errfunc(err)
	defer file.Close()
	for {
		fmt.Print("Введите № требуемого архива: ")
		if fmt.Scanln(&needArch); needArch == 0 {
			needArch = 1073741824 //Запрос последнего архива. Если запрашиваемый №архива > №последнего - ПЛК вышлет последний.
		}
		// Формирование и отправка запроса
		request(Conf.Plc, needArch, file)
		// Чтение ответа
		RTMresponse := Read_response(file)
		fmt.Println(RTMresponse)
		//Разбор шаки архива
		var t = 9 //адрес начала данных в ответе
		for i, count := range Conf.TitleBytes {
			for j := 0; j < count; j++ {
				if j == 0 {
					fmt.Println(RTMresponse[t], "---", Conf.TitleBytes[i], "---", Conf.TitleSpec[i])
				} else {
					fmt.Println(RTMresponse[t])
				}
				t++
			}
		}
	}

}
