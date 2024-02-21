package main

// Система учета топлива КВАРТА-Р1 (блок учета топлива БУТ-Р1)
// обязательно завершать скрипт кнопкой "Остановить", иначе процесс висит

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/annettalekto/ipkwiz"

	"github.com/amdf/ixxatvci3"
	"github.com/amdf/ixxatvci3/candev"
	"gopkg.in/ini.v1"
)

var wiz ipkwiz.Wizard
var can25 *candev.Device

func main() {

	wiz.Init()
	if !wiz.Open() {
		fmt.Println("Нет связи с МС")
		return
	}
	defer wiz.Close()
	defer wiz.Stop()

	//регистрируем падение скрипта
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			fmt.Printf("У скрипта паника %c%c%c", 128561, 128561, 128561)
			wiz.Error("Аварийное завершение")
			wiz.Stop()
			os.Exit(1)
		}
	}()

	initCAN25()
	go canStatusThread()

	iniScriptName := strings.TrimSuffix(filepath.Base(os.Args[0]), `.exe`) + `.ini`
	curDir, _ := os.Getwd()
	iniFullPath := curDir + `\` + iniScriptName
	// cfg, err := ini.Load(iniFullPath)

	wiz.Title("Эмулятор КВАРТА-Р1")
	wiz.Msg(" ")

	wiz.Msg("Эмуляция сообщений...")

	tankNumber := 4 // количество баков
	tabNumber := 0
	dataStatus := byte(0x1F)           // значение всех параметров действительно
	error1, error2 := byte(0), byte(0) // нет ошибок
	var iVol, iVol20, iMass, iTemp, iDens int

	for !wiz.NeedStop() {

		cfg, err := ini.Load(iniFullPath)
		if err != nil {
			wiz.Error(fmt.Sprintf("Ошибка чтения ini файла: %s, err = %v"+iniFullPath, err))
			return
		}

		v1, v2, ok := getDataBUTR(cfg)
		if ok {
			tankNumber = v1

			// Табельный номер машиниста для БУТ-Р1. Отсылается после установки табельного номера машиниста
			if v2 != tabNumber { // по изменению
				tabNumber = v2
				sendTubNumber(tabNumber)
				wiz.Action(fmt.Sprintf("Табельный номер: %d", tabNumber))
				time.Sleep(4 * time.Second)
			}
		}

		// Слово состояния БУТ-Р1
		val1, val2, val3, ok := getStatusFromINI(cfg)
		if ok {
			dataStatus, error1, error2 = val1, val2, val3
		}

		sendStatusBUTR1(1, dataStatus, error1, error2)
		wiz.Action(fmt.Sprintf("Статус (бак 1) данные: %08b; ошибки: %08b, %08b", dataStatus, error1, error2))
		time.Sleep(200 * time.Millisecond)
		if 2 <= tankNumber {
			sendStatusBUTR1(2, dataStatus, error1, error2)
			wiz.Action(fmt.Sprintf("Статус (бак 2) данные: %08b; ошибки: %08b, %08b", dataStatus, error1, error2))
			time.Sleep(200 * time.Millisecond)
		}
		if 3 <= tankNumber {
			sendStatusBUTR1(3, dataStatus, error1, error2)
			wiz.Action(fmt.Sprintf("Статус (бак 3) данные: %08b; ошибки: %08b, %08b", dataStatus, error1, error2))
			time.Sleep(200 * time.Millisecond)
		}
		if 4 == tankNumber {
			sendStatusBUTR1(4, dataStatus, error1, error2)
			wiz.Action(fmt.Sprintf("Статус (бак 4) данные: %08b; ошибки: %08b, %08b", dataStatus, error1, error2))
			time.Sleep(200 * time.Millisecond)
		}

		time.Sleep(4 * time.Second)

		// Данные БУТ-Р1
		v1, v2, v3, v4, v5, ok := getDataFromINI(cfg)
		if ok {
			iVol, iVol20, iMass, iTemp, iDens = v1, v2, v3, v4, v5
		}

		sendDataBUTR1(1, iVol, iVol20, iMass, iTemp, iDens)
		wiz.Action(fmt.Sprintf("Данные (бак 1): объём %d л, объём (20°C) %d л, масса %d кг, темп. %d°C, плотность %d кг/м3", iVol, iVol20, iMass, iTemp, iDens))
		time.Sleep(200 * time.Millisecond)
		if 2 <= tankNumber {
			sendDataBUTR1(2, iVol, iVol20, iMass, iTemp, iDens)
			wiz.Action(fmt.Sprintf("Данные (бак 2): объём %d л, объём (20°C) %d л, масса %d кг, темп. %d°C, плотность %d кг/м3", iVol, iVol20, iMass, iTemp, iDens))
			time.Sleep(200 * time.Millisecond)
		}
		if 3 <= tankNumber {
			sendDataBUTR1(3, iVol, iVol20, iMass, iTemp, iDens)
			wiz.Action(fmt.Sprintf("Данные (бак 3): объём %d л, объём (20°C) %d л, масса %d кг, темп. %d°C, плотность %d кг/м3", iVol, iVol20, iMass, iTemp, iDens))
			time.Sleep(200 * time.Millisecond)
		}
		if 4 == tankNumber {
			sendDataBUTR1(4, iVol, iVol20, iMass, iTemp, iDens)
			wiz.Action(fmt.Sprintf("Данные (бак 4): объём %d л, объём (20°C) %d л, масса %d кг, темп. %d°C, плотность %d кг/м3", iVol, iVol20, iMass, iTemp, iDens))
			time.Sleep(200 * time.Millisecond)
		}

		time.Sleep(4 * time.Second)
	}
}

func initCAN25() {
	var err error

	var b candev.Builder
	can25, err = b.Speed(ixxatvci3.Bitrate25kbps).Mode("11bit").SelectDevice(true).Get()
	if err != nil {
		wiz.Error("Не удалось произвести инициализацию CAN 25, err=" + err.Error())
		os.Exit(0)
	}
	can25.Run()
}

func canStatusThread() {
	for !wiz.NeedStop() {
		wiz.SendCanBusLoad(can25.GetBusLoad(1))
		time.Sleep(time.Millisecond * 200)
	}

	can25.Stop()
	os.Exit(1)
}

// Получить данные БУТ-Р1 из ini
func getDataBUTR(cfg *ini.File) (iTankNumber, iTabNumber int, ok bool) {
	PROP := "Properties"
	sTankNumber := "Количество баков"
	sTabNumber := "Табельный номер"
	ok = true

	iTankNumber, err := cfg.Section(PROP).Key(sTankNumber).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sTankNumber, err.Error()))
		ok = false
	} else if 1 > iTankNumber || 4 < iTankNumber {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен от 1 до 4", sTankNumber))
		ok = false
	}

	iTabNumber, err = cfg.Section(PROP).Key(sTabNumber).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sTabNumber, err.Error()))
		ok = false
	}

	return
}

// получить состояние данных из ini - 1й байт статуса
// 1 - значение параметра действительно; 0 - значение параметра не действительно
/*func getDataStatusFromINI(cfg *ini.File) (status byte, ok bool) { // очень длинно и не нужно
	PROP := "Properties"
	ok = true
	status = 0x1F // значение всех параметров действительно

	sVol := "Значение действительно (объём топлива)"
	sVol20 := "Значение действительно (объём топлива при 20°C)"
	sMass := "Значение действительно (масса топлива)"
	sTemp := "Значение действительно (температура топлива)"
	sDens := "Значение действительно (плотность топлива)"

	iVol, err := cfg.Section(PROP).Key(sVol).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sVol, err.Error()))
		ok = false
	} else if 0 != iVol && 1 != iVol {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 1 (значение параметра действительно) или 0 (значение параметра не действительно)", sVol))
		ok = false
	} else {
		if 0 == iVol {
			status &= 0xFE
		} else {
			status |= 0x01
		}
	}

	iVol20, err := cfg.Section(PROP).Key(sVol20).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sVol20, err.Error()))
		ok = false
	} else if 0 != iVol20 && 1 != iVol20 {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 1 (значение параметра действительно) или 0 (значение параметра не действительно)", sVol20))
		ok = false
	} else {
		if 0 == iVol20 {
			status &= 0xFD
		} else {
			status |= 0x02
		}
	}

	iMass, err := cfg.Section(PROP).Key(sMass).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sMass, err.Error()))
		ok = false
	} else if 0 != iMass && 1 != iMass {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 1 (значение параметра действительно) или 0 (значение параметра не действительно)", sMass))
		ok = false
	} else {
		if 0 == iMass {
			status &= 0xFB
		} else {
			status |= 0x04
		}
	}

	iTemp, err := cfg.Section(PROP).Key(sTemp).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sTemp, err.Error()))
		ok = false
	} else if 0 != iTemp && 1 != iTemp {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 1 (значение параметра действительно) или 0 (значение параметра не действительно)", sTemp))
		ok = false
	} else {
		if 0 == iTemp {
			status &= 0xF7
		} else {
			status |= 0x08
		}
	}

	iDens, err := cfg.Section(PROP).Key(sDens).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sDens, err.Error()))
		ok = false
	} else if 0 != iDens && 1 != iDens {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 1 (значение параметра действительно) или 0 (значение параметра не действительно)", sDens))
		wiz.Msg(" ")
		ok = false
	} else {
		if 0 == iDens {
			status &= 0xEF
		} else {
			status |= 0x10
		}
	}

	return
}*/

// Получить статус из ini
// status - 1 байт словосостояния БУТ-Р1, состояние данных (1 - значение парам. действительно)
// error - 2 и 3 байты словосостояния. Значение бита: 1 – ошибка есть, 0 – ошибки нет.
func getStatusFromINI(cfg *ini.File) (status, error1, error2 byte, ok bool) {
	PROP := "Properties"
	ok = true

	status = 0x1F   // значение всех параметров действительно
	sH550 := "Н550" //(система не отвечает);
	sH551 := "Н551" //(неисправен датчик температуры №1);
	sH552 := "Н552" //(неисправен датчик температуры №2);
	sH553 := "Н553" //(неисправен датчик уровня №1);
	sH554 := "Н554" //(неисправен датчик уровня №2);
	sH555 := "Н555" //(неисправен датчик плотности);
	sH556 := "Н556" //(некорректные значения начальной плотности и температуры);
	sH557 := "Н557" //(не удалось записать начальные значения плотности и температуры);
	sH560 := "Н560" //(нет связи с РПЗУ Измерителя);
	sH561 := "Н561" //(сбой информации в РПЗУ Измерителя);
	sH562 := "Н562" //(нет связи с РПЗУ Регистратора);
	sH563 := "Н563" //(сбой информации в РПЗУ Регистратора);
	sH564 := "Н564" //(нет связи с часами реального времени);
	sH565 := "Н565" //(нет связи с кнопками);
	sH566 := "Н566" //(ошибка CAN);
	sH608 := "Н608" //(память заполнена более чем на 90%);

	iVal, err := cfg.Section(PROP).Key(sH550).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH550, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH550))
		ok = false
	} else {
		if 0 == iVal {
			error1 &= 0xFE
		} else {
			error1 |= 0x01
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH551).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH551, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH551))
		ok = false
	} else {
		if 0 == iVal {
			error1 &= 0xFD
		} else {
			error1 |= 0x02
			status &= 0xFE // значение темп недействительно
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH552).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH552, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH552))
		ok = false
	} else {
		if 0 == iVal {
			error1 &= 0xFB
		} else {
			error1 |= 0x04
			status &= 0xFE // значение темп недействительно
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH553).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH553, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH553))
		ok = false
	} else {
		if 0 == iVal {
			error1 &= 0xF7
		} else {
			error1 |= 0x08
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH554).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH554, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH554))
		ok = false
	} else {
		if 0 == iVal {
			error1 &= 0xEF
		} else {
			error1 |= 0x10
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH555).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH555, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH555))
		ok = false
	} else {
		if 0 == iVal {
			error1 &= 0xDF
		} else {
			error1 |= 0x20
			status &= 0xEF // значение плотности недействительно
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH556).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH556, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH556))
		ok = false
	} else {
		if 0 == iVal {
			error1 &= 0xBF
		} else {
			error1 |= 0x40
			status &= 0xFE // значение темп недействительно
			status &= 0xEF // значение плотности недействительно
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH557).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH557, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH557))
		ok = false
	} else {
		if 0 == iVal {
			error1 &= 0x7F
		} else {
			error1 |= 0x80
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH560).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH560, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH560))
		ok = false
	} else {
		if 0 == iVal {
			error2 &= 0xFE
		} else {
			error2 |= 0x01
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH561).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH561, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH561))
		ok = false
	} else {
		if 0 == iVal {
			error2 &= 0xFD
		} else {
			error2 |= 0x02
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH562).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH562, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH562))
		ok = false
	} else {
		if 0 == iVal {
			error2 &= 0xFB
		} else {
			error2 |= 0x04
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH563).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH563, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH563))
		ok = false
	} else {
		if 0 == iVal {
			error2 &= 0xF7
		} else {
			error2 |= 0x08
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH564).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH564, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH564))
		ok = false
	} else {
		if 0 == iVal {
			error2 &= 0xEF
		} else {
			error2 |= 0x10
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH565).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH565, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH565))
		ok = false
	} else {
		if 0 == iVal {
			error2 &= 0xDF
		} else {
			error2 |= 0x20
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH566).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH566, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH566))
		ok = false
	} else {
		if 0 == iVal {
			error2 &= 0xBF
		} else {
			error2 |= 0x40
		}
	}

	iVal, err = cfg.Section(PROP).Key(sH608).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sH608, err.Error()))
		ok = false
	} else if 0 != iVal && 1 != iVal {
		wiz.Error(fmt.Sprintf("Параметр: \"%s\" должен быть равен 0 (нет ошибки) или 1 (ошибка)", sH608))
		wiz.Msg(" ")
		ok = false
	} else {
		if 0 == iVal {
			error2 &= 0x7F
		} else {
			error2 |= 0x80
		}
	}

	return
}

// получить состояние данных из ini - 1й байт статуса
func getDataFromINI(cfg *ini.File) (iVol, iVol20, iMass, iTemp, iDens int, ok bool) {
	PROP := "Properties"
	ok = true

	sVol := "Объём топлива"
	sVol20 := "Объём топлива при 20°C"
	sMass := "Масса топлива"
	sTemp := "Температура топлива"
	sDens := "Плотность топлива"

	iVol, err := cfg.Section(PROP).Key(sVol).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sVol, err.Error()))
		ok = false
	}

	iVol20, err = cfg.Section(PROP).Key(sVol20).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sVol20, err.Error()))
		ok = false
	}

	iMass, err = cfg.Section(PROP).Key(sMass).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sMass, err.Error()))
		ok = false
	}

	iTemp, err = cfg.Section(PROP).Key(sTemp).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sTemp, err.Error()))
		ok = false
	}

	iDens, err = cfg.Section(PROP).Key(sDens).Int()
	if err != nil {
		wiz.Error(fmt.Sprintf("Не получено: \"%s\", err=%v", sDens, err.Error()))
		wiz.Msg(" ")
		ok = false
	}

	return
}

// Табельный номер машиниста для БУТ-Р1. Отсылается после установки табельного номера машиниста (те по изменению)
func sendTubNumber(tubNumber int) (ok bool) {
	var msg candev.Message
	ok = true

	msg.ID = 0x3C3
	msg.Len = 2
	msg.Data[0] = byte(tubNumber)
	msg.Data[1] = byte(tubNumber >> 8)

	can25.Send(msg)

	return
}

func sendStatusBUTR1(tankNumber int, status, error1, error2 byte) (ok bool) {
	var msg candev.Message
	ok = true

	id := uint32(0x3C0) // id для первого бака
	msg.ID = id + (uint32(tankNumber)-1)*0x10
	msg.Len = 8
	msg.Data[0] = status
	msg.Data[1] = error1
	msg.Data[2] = error2

	can25.Send(msg)

	return
}

func sendDataBUTR1(tankNumber, iVol, iVol20, iMass, iTemp, iDens int) (ok bool) {
	var msg candev.Message
	ok = true
	const lowerDens int = 840

	id := uint32(0x3C1) // id для первого бака
	msg.ID = id + (uint32(tankNumber)-1)*0x10
	msg.Len = 8
	msg.Data[0] = byte(iVol)      // todo узнать в каких единицах отправлять данные, предпологаю:
	msg.Data[1] = byte(iVol >> 8) // мм
	msg.Data[2] = byte(iVol20)
	msg.Data[3] = byte(iVol20 >> 8)
	msg.Data[4] = byte(iMass) // кг
	msg.Data[5] = byte(iMass >> 8)
	msg.Data[6] = byte(iTemp)
	msg.Data[7] = byte(iDens - lowerDens) // кг/м3 (вычитая 840)

	can25.Send(msg)

	return
}
