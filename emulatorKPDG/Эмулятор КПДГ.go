package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mjk712/ipkEmulator/kpdg"

	common "github.com/annettalekto/ipkCmn"
	"github.com/annettalekto/ipkwiz"

	"github.com/amdf/ixxatvci3/candev"

	"gopkg.in/ini.v1"
)

var wiz ipkwiz.Wizard

func main() {
	// scriptVersion := "0.0.1"
	var can25 candev.Device
	var data kpdg.DataType

	wiz.Init()
	if !wiz.Open() {
		fmt.Printf("%c Нет связи с МС %c", 128561, 128542)
		return
	}
	defer wiz.Close()
	defer wiz.Stop()

	// регистрируем падение скрипта
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("У скрипта паника %c%c%c", 128561, 128561, 128561)
			wiz.Error("Аварийное завершение")
			wiz.Stop()
			os.Exit(1)
		}
	}()

	err := can25.Init(0x1F, 0x16)
	if err != nil {
		wiz.Error(fmt.Sprintf("Не удалось произвести инициализацию CAN"))
		return
	}
	can25.Run()
	defer can25.Stop()

	// ini
	nameCheck := strings.TrimSuffix(filepath.Base(os.Args[0]), `.exe`)
	if nameCheck == "__debug_bin" {
		nameCheck = "Эмулятор КПДГ"
	}
	curDir, _ := os.Getwd()
	iniPath := curDir + `\` + nameCheck + `.ini`
	cfg, err := ini.Load(iniPath)
	if err != nil {
		wiz.Error(fmt.Sprintf("Ошибка чтения ini файла \"%s\": %v", iniPath, err))
		return
	}
	data, err = getProperties(cfg)

	// запуск эмуляции
	var kpdg1 kpdg.Emulator
	kpdg1.Init(&can25, data)
	kpdg1.Start()
	defer kpdg1.Stop()

	wiz.Title("Эмуляция КПДГ")

	wiz.Msg(" ")
	if data.EmulateSD {
		wiz.Action("Имитация собщений CD-карты")
		time.Sleep(time.Second)
	}

	for !wiz.NeedStop() {
		wiz.Action("Имитация данных КРПД")
		time.Sleep(time.Second)
		wiz.Action("Имитация частотных данных КРПД")
		time.Sleep(time.Second)
		wiz.Action("Имитация аналоговых данных КРПД")
		time.Sleep(time.Second)
		wiz.Action("Имитация словосостояния КРПД")
		time.Sleep(time.Second)
	}
}

const (
	iniKrpdNum      = "Номер КРПД в сцепке"
	iniErrorCode1   = "Код ошибки 1"
	iniErrorCode2   = "Код ошибки 2"
	iniErrorCode3   = "Код ошибки 3"
	iniErrorCode4   = "Код ошибки 4"
	iniKmPos        = "Позиция контроллера машиниста"
	iniVoltage      = "Напряжение на зажимах тягового генератора (В)"
	iniCurrent      = "Ток тягового генератора (А)"
	iniOilPressure  = "Давление масла в масляной системе дизеля"
	iniFuelPressure = "Давление топлива в коллекторе низкого давления"
	iniOilTemp      = "Температура масла на выходе из дизеля (°С)"
	iniWaterTemp    = "Температура воды в горячем контуре системы охлаждения (°С)"
	iniAirPressure  = "Давление воздуха в воздушном ресивере турбонаддува"
	iniFreq         = "Частота вращения вала дизеля (об/мин)"
	iniImitateSD    = "Имитация замены SD карты"
	iniPressureUni  = "Единица измерения давления"
)

func getProperties(cfg *ini.File) (v kpdg.DataType, err error) {
	PROP := "Properties"
	var ival int

	var pressureUnitKPA bool // давление задавали в кПа?
	str := cfg.Section(PROP).Key(iniPressureUni).String()
	if str != "0" && str != "1" {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniPressureUni, str, err))
	} else if str == "1" {
		pressureUnitKPA = true
	}

	str = cfg.Section(PROP).Key(iniKrpdNum).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniKrpdNum, str, err))
	} else {
		v.NumberKPDG = ival
	}

	str = cfg.Section(PROP).Key(iniErrorCode1).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniErrorCode1, str, err))
	} else {
		v.ErrorCode1 = ival
	}

	str = cfg.Section(PROP).Key(iniErrorCode2).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniErrorCode2, str, err))
	} else {
		v.ErrorCode2 = ival
	}

	str = cfg.Section(PROP).Key(iniErrorCode3).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniErrorCode3, str, err))
	} else {
		v.ErrorCode3 = ival
	}

	str = cfg.Section(PROP).Key(iniErrorCode4).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniErrorCode4, str, err))
	} else {
		v.ErrorCode4 = ival
	}

	str = cfg.Section(PROP).Key(iniKmPos).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniKmPos, str, err))
	} else {
		v.PositionKM = ival
	}

	str = cfg.Section(PROP).Key(iniVoltage).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniVoltage, str, err))
	} else {
		v.Voltage = ival
	}

	str = cfg.Section(PROP).Key(iniCurrent).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniCurrent, str, err))
	} else {
		v.Current = ival
	}

	str = cfg.Section(PROP).Key(iniOilPressure).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniOilPressure, str, err))
	} else {
		if pressureUnitKPA {
			fval := common.KiloPascalToAt(float64(ival))
			ival = int(fval * 1000)
		}
		v.OilPressure = ival // давление в кгс/см2

	}

	str = cfg.Section(PROP).Key(iniFuelPressure).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniFuelPressure, str, err))
	} else {
		if pressureUnitKPA {
			fval := common.KiloPascalToAt(float64(ival))
			ival = int(fval * 1000)
		}
		v.FuelPressure = ival
	}

	str = cfg.Section(PROP).Key(iniOilTemp).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniOilTemp, str, err))
	} else {
		v.OilTemp = ival
	}

	str = cfg.Section(PROP).Key(iniWaterTemp).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniWaterTemp, str, err))
	} else {
		v.WaterTemp = ival
	}

	str = cfg.Section(PROP).Key(iniAirPressure).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniAirPressure, str, err))
	} else {
		if pressureUnitKPA {
			fval := common.KiloPascalToAt(float64(ival))
			ival = int(fval * 1000)
		}
		v.AirPressure = ival
	}

	str = cfg.Section(PROP).Key(iniFreq).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniFreq, str, err))
	} else {
		v.Freq = ival
	}

	str = cfg.Section(PROP).Key(iniImitateSD).String()
	if ival, err = strconv.Atoi(str); err != nil {
		wiz.Error(fmt.Sprintf("Неверное значение параметра «%s»: %s (err: %v)", iniImitateSD, str, err))
	} else if ival == 1 {
		v.EmulateSD = true
	}

	return
}
