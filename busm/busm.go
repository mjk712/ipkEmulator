package busm

import (
	"errors"
	"fmt"
	"time"

	"github.com/amdf/ixxatvci3/candev"
	"github.com/annettalekto/ipkwiz"
)

var can25 *candev.Device
var wiz *ipkwiz.Wizard

// заметки: ключ эпк должен быть включен при проверке АЛС (ош 152)

// InitEmulator начать
func InitEmulator(c *candev.Device, w *ipkwiz.Wizard) (err error) {
	if nil == w || nil == c {
		err = errors.New("InitEmulator(): parameter = nil, err: %v" + err.Error())
		return
	}

	can25 = c
	wiz = w

	info.busStatus = 0
	info.busKm = S395_OTPUSK
	info.busCock395 = S395_OTPUSK
	info.busInputs = BUS_INP_CAB1 | BUS_INP_EPKKEY

	return
}

// Emulate продолжтть
func Emulate() {

	if !info.emulation {
		go func() {
			fmt.Println("Запуск эмуляции БУС-М")
			info.emulation = true

			for info.emulation {
				sendStatus()
				sendData()
				sendFireAlarm()
				time.Sleep(time.Second)
			}
		}()
	}
}

// Stop не продолжать
func Stop() {
	info.emulation = false
}

//---------------------------------------------------------------------
//                               Этапы
//---------------------------------------------------------------------

// заметки БУС: одна кабина должна быть всегда (ош 181)

// ModeBUS режим БУС
func ModeBUS() {

	allInputs := []int{BUS_INP_CLEAT0, BUS_INP_TSKBM, BUS_INP_EPKKEY, BUS_INP_EPKCOCK1, BUS_INP_EPKCOCK2,
		BUS_INP_D3, BUS_INP_EPK2, BUS_INP_SAUT, BUS_INP_KATEPK1, BUS_INP_KATEPK2 /*, BUS_INP_CAB1, BUS_INP_CAB2*/}

	for i := 0; i <= 1; i++ { // дважды выдаем все
		for _, inputs := range allInputs {
			info.busInputs = uint16(BUS_INP_CAB2 | inputs)
			time.Sleep(5 * time.Second)
		}
	}
}

// ModeKKM режим ККМ
func ModeKKM() {

	allStates395 := []int{S395_OTPUSK, S395_PERE, S395_TORM, S395_EXTR_TORM, S395_PERE_SAUT, S395_TORM_SAUT}

	for i := 0; i <= 1; i++ { // дважды выдаем все
		for _, state := range allStates395 {
			info.busKm = uint16(state)
			info.busCock395 = uint16(state)
			time.Sleep(5 * time.Second)
		}
	}
}

// ModeFireAlarm режим ПС
func ModeFireAlarm() {

	allStates := []uint{
		// 1 кабина(1 байт: 0 бит -> 3 бит): ПС ВКЛ=1,  ПС АВТ=1, 	ПС не АКТ=0, СПТ АКТ=1	(=0B)
		// 2 кабина(2 байт: 0 бит -> 3 бит): ПС ВЫКЛ=0, ПС АВТ=1, 	ПС АКТ=1,    СПТ не АКТ=0 (=06)
		0x060B,
		// 1 кабина(1 байт: 0 бит -> 3 бит): ПС ВЫКЛ=0, ПС РУЧНОЙ=0, ПС АКТ=1, 	  СПТ не АКТ=0 (=04)
		// 2 кабина(2 байт: 0 бит -> 3 бит): ПС ВКЛ=1,  ПС РУЧНОЙ=0, ПС не АКТ=0, СПТ АКТ=1 (=09)
		0x0904,
	}

	for i := 0; i <= 1; i++ { // дважды выдаем все
		for _, state := range allStates {
			info.fireAlarm = uint16(state)
			time.Sleep(5 * time.Second)
		}
	}
}
