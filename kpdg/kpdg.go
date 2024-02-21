package kpdg

import (
	"errors"
	"fmt"
	"ipkwiz"
	"time"

	"github.com/amdf/ixxatvci3/candev"
)

var can25 *candev.Device
var wiz *ipkwiz.Wizard

// Init начать
func (e *Emulator) Init(c *candev.Device /*, w *ipkwiz.Wizard*/, v DataType) (err error) {
	if /*nil == w ||*/ nil == c {
		err = errors.New("Init(): parameter = nil, err: %v" + err.Error())
		return
	}

	can25 = c
	// wiz = w

	// проверить бы данные
	if v.NumberKPDG < 1 || v.NumberKPDG > 4 {
		err = errors.New("InitEmulator(): в сети может использоваться не более четырёх устройств КРПД")
	}

	e.Data = v

	return
}

// Start продолжить
func (e *Emulator) Start() {

	if !e.Enabled {
		go func() {
			fmt.Printf("Запуск эмуляции КПДГ %d", e.Data.NumberKPDG)
			e.Enabled = true

			if e.Data.EmulateSD {
				e.sendInSD()
				e.sendOutSD()
			}
			e.sendData()

			for e.Enabled {

				e.sendFrqData()
				e.sendAnlData()
				e.sendData()
				e.sendStatus()
				time.Sleep(1 * time.Second)
			}
		}()
	}
}

// Stop останавливает запущенную эмуляцию.
func (e *Emulator) Stop() {
	fmt.Printf("Эмуляции КПДГ %d СТОП", e.Data.NumberKPDG)

	e.Enabled = false
}
