// Code generated .* DO NOT EDIT.
package busm

import (
	"fmt"
	"time"

	"github.com/amdf/ixxatvci3/candev"
)

const (
	ID_BUS_DATA   = 0x592
	ID_ZAPROS_BUS = 0x591 // Слово состояние БУС-М
	ID_KKM_DATA1  = 0x570
	ID_KKM_DATA2  = 0x571
	ID_DATA_BUSMP = 0x597 // Данные БУС-М (сигналы пожарной безопасности)
)

const (
	BUS_INP_CLEAT0   = 0x0001 ///< Клемма "0" ЭПТ
	BUS_INP_TSKBM    = 0x0002 ///< ТСКБМ
	BUS_INP_EPKKEY   = 0x0004 ///< Ключ ЭПК
	BUS_INP_EPKCOCK1 = 0x0008 ///< Кран ЭПК (кабина 1)
	BUS_INP_EPKCOCK2 = 0x0010 ///< Кран ЭПК (кабина 2)
	BUS_INP_D3       = 0x0020 ///< Тумблер Д3
	BUS_INP_EPK2     = 0x0040 ///< ЭПК2
	BUS_INP_SAUT     = 0x0080 ///< САУТ ? реле ИФ
	BUS_INP_KATEPK1  = 0x0100 ///< Катушка ЭПК (кабина 1)
	BUS_INP_KATEPK2  = 0x0200 ///< Катушка ЭПК (кабина 2)
	BUS_INP_CAB1     = 0x0400 ///< Ведущая кабина 1
	BUS_INP_CAB2     = 0x0800 ///< Ведущая кабина 2
)

// кодирование контактов крана 395
const (
	P395_EXTR   = 0x08
	P395_TORM   = 0x04
	P395_PERE   = 0x02
	P395_OTPUSK = 0x01
)

// стандартные положения крана
const (
	S395_OTPUSK    = P395_OTPUSK | P395_EXTR
	S395_PERE      = P395_PERE | P395_EXTR
	S395_TORM      = P395_TORM | P395_EXTR
	S395_EXTR_TORM = P395_TORM
	S395_PERE_SAUT = P395_OTPUSK | P395_PERE | P395_EXTR
	S395_TORM_SAUT = P395_OTPUSK | P395_TORM | P395_EXTR
)

type data struct {
	busStatus  byte   // словосостояние
	busKm      uint16 // состояние контроллера машиниста
	busCock395 uint16 // положение крана 395 1 и 2 каб
	busInputs  uint16 // сигн
	fireAlarm  uint16 // данные пожарной безопасности
	emulation  bool   // начать/остановить эмуляцию
}

var info data

func sendStatus() (err error) {
	var msg candev.Message

	msg.ID = uint32(ID_ZAPROS_BUS)
	msg.Len = 3

	msg.Data = [8]byte{byte(info.busStatus), 0, 0}
	err = can25.Send(msg)

	// fmt.Printf("Словосостояние БУС-М: %X\n", msg.Data)
	time.Sleep(200 * time.Millisecond)

	return
}

func sendData() (err error) {
	var msg candev.Message
	msg.ID = uint32(ID_BUS_DATA)
	msg.Len = 6

	mod := byte(0x01) // модификатор состояние двоичных входов
	km1 := byte(info.busKm)
	km2 := byte(info.busKm << 8)
	position395 := byte(info.busCock395)
	inputs1 := byte(info.busInputs)
	inputs2 := byte(info.busInputs >> 8)

	msg.Data = [8]byte{mod, km1, km2, (position395<<4 | position395), inputs1, inputs2}
	err = can25.Send(msg)

	fmt.Printf("Данные БУС-М: %X\n", msg.Data)
	time.Sleep(200 * time.Millisecond)

	return
}

func sendFireAlarm() (err error) {
	var msg candev.Message

	if info.fireAlarm == 0 {
		return
	}

	msg.ID = uint32(ID_DATA_BUSMP)
	msg.Len = 2

	msg.Data = [8]byte{byte(info.fireAlarm), byte(info.fireAlarm >> 8)}
	err = can25.Send(msg)

	fmt.Printf("ПС: %X\n", msg.Data)
	time.Sleep(200 * time.Millisecond)

	return
}
