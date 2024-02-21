package kpdg

import (
	"fmt"
	"time"

	"github.com/amdf/ixxatvci3/candev"
)

// DataType данные для имитации
type DataType struct {
	NumberKPDG   int  // "Номер КРПД в сцепке": 1-4
	ErrorCode1   int  // код ошибки 1: 350-357
	ErrorCode2   int  // 358-365
	ErrorCode3   int  // 366-373
	ErrorCode4   int  // 374-381
	PositionKM   int  // Позиция контроллера машиниста
	Freq         int  // Частота вращения вала дизеля (об/мин)
	Voltage      int  // Напряжение на зажимах тягового генератора (В)
	Current      int  // Ток тягового генератора (А)
	OilPressure  int  // Давление масла в масляной системе дизеля (кгс/см2)			! если в ini были кПА, перевести заранее!
	FuelPressure int  // Давление топлива в коллекторе низкого давления (кгс/см2) 	! если в ini были кПА, перевести заранее!
	OilTemp      int  // Температура масла на выходе из дизеля (°С)
	WaterTemp    int  // Температура воды в горячем контуре системы охлаждения (°С)
	AirPressure  int  // Давление воздуха в воздушном ресивере турбонаддува	(кгс/см2)! если в ini были кПА, перевести заранее!
	EmulateSD    bool // Имитировать собщения CD-карты
}

// Emulator для эмуляции
type Emulator struct {
	Enabled bool
	Data    DataType
}

// number -- Номер КРПД в сцепке
func getID(number int) (id uint32) {
	ids := []int{0x6AC, 0x6AD, 0x6AE, 0x6AF}
	id = uint32(ids[number])
	return
}

/*
KPDG_STATUS_1 6ACH 6ADH 6AEH 6AFH
Слово состояния КРПД (в сети может использоваться до четырёх устройств КРПД)
Слово состояния КПДГ. Периодичность отправки 3 с. (2 байт)
1 байт – 0x00, модификатор команды.
2 байт – техническое состояние КРПД, 0 – исправен, 1 – неисправен.
Бит 0 – исправность КПДГ;
Бит 1 – исправность датчика давления топлива;
Бит 2 – исправность датчика давления масла;
Бит 3 – исправность датчика температуры масла;
Бит 4 – исправность датчика температуры охл. жидк.;
Бит 5 – исправность датчика частоты вращения вала;
Бит 6 – исправность датчика давления наддува;
Бит 7 – исправность шунта.
*/

func (e *Emulator) getStatus() (st byte) {
	errors := map[int]byte{ // 350-357
		0x15E: 0x01, // неисправность КПДГ
		0x15F: 0x02, // неисправность датчика давления топлива
		0x160: 0x04, // неисправность датчика давления масла
		0x161: 0x08, // неисправность датчика температуры масла
		0x162: 0x10, // неисправность датчика температуры охл. жидк
		0x163: 0x20, // неисправность датчика частоты вращения вала
		0x164: 0x40, // неисправность датчика давления наддува
		0x165: 0x80, // неисправность шунта
	}
	st |= errors[e.Data.ErrorCode1]
	st |= errors[e.Data.ErrorCode2]
	st |= errors[e.Data.ErrorCode3]
	st |= errors[e.Data.ErrorCode4]

	return
}

// Слово состояния КРПД, 3 сек
func (e *Emulator) sendStatus() (err error) {
	var msg candev.Message

	mod := 0
	status := e.getStatus()
	msg.ID = getID(e.Data.NumberKPDG - 1)
	msg.Len = 2
	msg.Data = [8]byte{byte(mod), byte(status)}
	err = can25.Send(msg)

	fmt.Printf("Словосостояние КРПД %d: %X, err: %v\n", e.Data.NumberKPDG, msg.Data, err)
	time.Sleep(200 * time.Millisecond)

	return
}

/*
KPDG_DATA_1 6ACH 6ADH 6AEH 6AFH
Данные КРПД
Данные КПДГ (длина 6 байт).
1 байт ¬– 0x01, модификатор команды.
2, 3 байты – позиция контроллера машиниста.
Байты 4 – 6 разбиты на идентичные полубайты, определяющие состояние измеряемого параметра:
0000 – параметр в норме;
0001 – выход параметра за нижний предел;
0010 – приближение параметра к нижнему пределу;
0011 – приближение параметра к верхнему пределу;
0100 – выход параметра за верхний предел.
4 байт – старший полубайт – частота вращения, младший – давление турбонаддува.
5 байт – старший полубайт – давление масла, младший – давление топлива.
6 байт – старший полубайт – температура масла, младший – температура воды.
*/

// pos -- Позиция контроллера машиниста
func getPosition(pos int) (p int) {
	positions := map[int]int{
		0: 0x00,
		1: 0x20,
		2: 0x30,
		3: 0x31,
		4: 0x33,
		5: 0x36,
		6: 0x39,
		7: 0x3C,
		8: 0x3F,
	}
	p = positions[pos]
	return
}

func (e *Emulator) getState1() (st byte) {
	errors := map[int]byte{ // 358-365
		0x166: 0x40, // выход частоты вращения за верхний предел
		0x167: 0x10, // выход частоты вращения за нижний предел
		0x168: 0x30, // приближение частоты вращения к верхнему пределу
		0x169: 0x20, // приближение частоты вращения к нижнему пределу
		0x16A: 0x04, // выход давления турбонаддува за верхний предел
		0x16B: 0x01, // выход давления турбонаддува за нижний предел
		0x16C: 0x03, // приближение давления турбонаддува к верхнему пределу
		0x16D: 0x02, // приближение давления турбонаддува к нижнему пределу
	}
	st |= errors[e.Data.ErrorCode1]
	st |= errors[e.Data.ErrorCode2]
	st |= errors[e.Data.ErrorCode3]
	st |= errors[e.Data.ErrorCode4]

	return
}

func (e *Emulator) getState2() (st byte) {
	errors := map[int]byte{ // 366-373
		0x16E: 0x40, // выход давления масла за верхний предел
		0x16F: 0x10, // выход давления масла за нижний предел
		0x170: 0x30, // приближение давления масла к верхнему пределу
		0x171: 0x20, // приближение давления масла к нижнему пределу
		0x172: 0x04, // выход давления топлива за верхний предел
		0x173: 0x01, // выход давления топлива за нижний предел
		0x174: 0x03, // приближение давления топлива к верхнему пределу
		0x175: 0x02, // приближение давления топлива к нижнему пределу
	}
	st |= errors[e.Data.ErrorCode1]
	st |= errors[e.Data.ErrorCode2]
	st |= errors[e.Data.ErrorCode3]
	st |= errors[e.Data.ErrorCode4]

	return
}

func (e *Emulator) getState3() (st byte) {
	errors := map[int]byte{ // 374-381
		0x176: 0x40, // выход температуры масла за верхний предел
		0x177: 0x10, // выход температуры масла за нижний предел
		0x178: 0x30, // приближение температуры масла к верхнему пределу
		0x179: 0x20, // приближение температуры масла к нижнему пределу
		0x17A: 0x04, // выход температуры воды за верхний предел
		0x17B: 0x01, // выход температуры воды за нижний предел
		0x17C: 0x03, // приближение температуры воды к верхнему пределу
		0x17D: 0x02, // приближение температуры воды к нижнему пределу
	}
	st |= errors[e.Data.ErrorCode1]
	st |= errors[e.Data.ErrorCode2]
	st |= errors[e.Data.ErrorCode3]
	st |= errors[e.Data.ErrorCode4]

	return
}

// sendData отправить данные КПДГ
func (e *Emulator) sendData() (err error) {
	var msg candev.Message

	mod := byte(0x01)
	pos := getPosition(e.Data.PositionKM) // todo ?
	state1 := e.getState1()
	state2 := e.getState2()
	state3 := e.getState3()

	msg.ID = getID(e.Data.NumberKPDG - 1)
	msg.Len = 6
	msg.Data = [8]byte{mod, byte(pos), byte(pos >> 8), byte(state1), byte(state2), byte(state3)}
	err = can25.Send(msg)

	fmt.Printf("Данные КРПД %d: %X\n", e.Data.NumberKPDG, msg.Data)
	time.Sleep(200 * time.Millisecond)

	return
}

/*
KPDG_FRQDATA_1 6ACH 6ADH 6AEH 6AFH
Частотные данные КРПД
Данные частотных входов КПДГ (длина 4 байта).
1 байт 0x02, модификатор команды.
2 байт – порядковый номер значения частотного входа.
«0x00» – частота вращения вала дизеля.
Байты 3..4:
значение частотного сигнала (длина 2 байта)
*/

func (e *Emulator) sendFrqData() (err error) {
	var msg candev.Message

	mod := byte(0x02) // модификатор
	number := byte(0)

	msg.ID = getID(e.Data.NumberKPDG - 1)
	msg.Len = 4
	msg.Data = [8]byte{mod, number, byte(e.Data.Freq), byte(e.Data.Freq >> 8)}
	err = can25.Send(msg)

	fmt.Printf("Частотные данные КРПД %d: %X\n", e.Data.NumberKPDG, msg.Data)
	time.Sleep(200 * time.Millisecond)

	return
}

/*
KPDG_ANLDATA_1 6ACH 6ADH 6AEH 6AFH
Аналоговые данные КРПД
Данные аналоговых входов КПДГ (длина 5 байт)
1 байт модификатор команды, 2, 3 байт – значение параметра, 4, 5 - значение параметра.
1 байт: модификатор команды «0x03»:
– Байты 2..3: напряжение тягового генератора;
– Байты 4..5: ток тягового генератора;
1 байт: модификатор команды «0x04»:
– Байты 2..3: давление масла дизеля;
– Байты 4..5: давление топлива.
1 байт: модификатор команды «0x05»:
– Байты 2..3: температура масла;
– Байты 4..5: температура воды.
1 байт: модификатор команды «0x06»:
– Байты 2..3: давление надувочного воздуха;
– Байты 4..5: зарезервировано.
*/

func (e *Emulator) sendAnlData() (err error) {
	var msg candev.Message
	msg.ID = getID(e.Data.NumberKPDG - 1)
	msg.Len = 5

	mod := byte(0x03)
	value1 := e.Data.Voltage
	value2 := e.Data.Current
	msg.Data = [8]byte{mod, byte(value1), byte(value1 >> 8), byte(value2), byte(value2 >> 8)}
	err = can25.Send(msg)
	fmt.Printf("Аналоговые данные 1 КРПД %d: %X\n", e.Data.NumberKPDG, msg.Data)
	time.Sleep(200 * time.Millisecond)

	mod = byte(0x04)
	value1 = e.Data.OilPressure
	value2 = e.Data.FuelPressure
	msg.Data = [8]byte{mod, byte(value1), byte(value1 >> 8), byte(value2), byte(value2 >> 8)}
	err = can25.Send(msg)
	fmt.Printf("Аналоговые данные 2 КРПД %d: %X\n", e.Data.NumberKPDG, msg.Data)
	time.Sleep(200 * time.Millisecond)

	mod = byte(0x05)
	value1 = e.Data.OilTemp
	value2 = e.Data.WaterTemp
	msg.Data = [8]byte{mod, byte(value1), byte(value1 >> 8), byte(value2), byte(value2 >> 8)}
	err = can25.Send(msg)
	fmt.Printf("Аналоговые данные 3 КРПД %d: %X\n", e.Data.NumberKPDG, msg.Data)
	time.Sleep(200 * time.Millisecond)

	mod = byte(0x06)
	value1 = e.Data.AirPressure
	value2 = 0
	msg.Data = [8]byte{mod, byte(value1), byte(value1 >> 8), byte(value2), byte(value2 >> 8)}
	err = can25.Send(msg)
	fmt.Printf("Аналоговые данные 4 КРПД %d: %X\n", e.Data.NumberKPDG, msg.Data)
	time.Sleep(200 * time.Millisecond)

	return
}

/*
KPDG_SDOUT_1 6ACH 6ADH 6AEH 6AFH
Сообщение об извлечении SD карты
1 байт модификатор команды «0x07».
Байты 2…5 – серийный номер извлеченной карты.
*/
func (e *Emulator) sendOutSD() (err error) {
	var msg candev.Message

	mod := byte(7)
	msg.ID = getID(e.Data.NumberKPDG - 1)
	msg.Len = 5
	msg.Data = [8]byte{mod, 0xF0, 0x15, 0xA5, 0xC7}
	err = can25.Send(msg)

	fmt.Printf("Сообщение об извлечении SD карты КРПД %d: %X\n", e.Data.NumberKPDG, msg.Data)
	time.Sleep(200 * time.Millisecond)

	return
}

/*
KPDG_SDIN_1 6ACH 6ADH 6AEH 6AFH
Сообщение об установке SD карты
1 байт модификатор команды «0x08».
Байты 2…5 – серийный номер установленной карты.
*/
func (e *Emulator) sendInSD() (err error) {
	var msg candev.Message

	mod := byte(0x08)
	v := 0x04030201
	msg.ID = getID(e.Data.NumberKPDG - 1)
	msg.Len = 5
	msg.Data = [8]byte{mod, byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
	err = can25.Send(msg)

	fmt.Printf("Сообщение об установке SD карты КРПД %d: %X\n", e.Data.NumberKPDG, msg.Data)
	time.Sleep(200 * time.Millisecond)

	return
}
