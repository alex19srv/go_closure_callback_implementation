package main

// как устроены callback при вызове замыкания

// доклад и слайды не мои
// доклад: https://www.youtube.com/watch?v=Vjo7Tmj3DnI
// слайды: https://docs.google.com/presentation/d/1jjajI3lwhZ9xYpKlPOuu3tmFmM-YXlrSpSHh2ay1ou0/edit#slide=id.g129ca4019df_1_161
// сгенерировать ассемблерный код на локальной машине: go build -gcflags -S x.go

import (
	"runtime"
)

type Void func()

var CBlist []Void // <- тут будут храниться не адреса функций, а адреса funcval,
//  в которых будут адреса функций и контекст, если фукнция - замыкание

// go:noinline
func AddCB(cb Void) {
	CBlist = append(CBlist, cb)
	// слайс состоит из указателя на буфер, количества элементов (в штуках) и емкости (в штуках)
	// MOVQ	main.CBlist+16(SB), CX <- емкость слайса
	// MOVQ	main.CBlist+8(SB), BX  <- текущий занятый индекс элемента
	// INCQ	BX                     <- следующий элемент
	// MOVQ	main.CBlist(SB), DX    <- указатель на буфер
	// тут был код проверки на переполнение слайса
	// MOVQ	AX, -8(DX)(BX*8)       <- по адресу DX + BX*8 записывается новый элемент
	// т.е. в коде нет разделения на обычные функции и замыкания
}

// go:noinline
func CallCBs() {
	// MOVQ	main.CBlist(SB), AX			<- указатель на слайс
	// MOVQ	AX, main..autotmp_4+16(SP)
	// MOVQ	main.CBlist+8(SB), CX		<- текущий занятый индекс элемента (конец итерации)
	// MOVQ	CX, main..autotmp_5+8(SP)
	// XORL	DX, DX						<- текущий индекс итерации DX = 0
	for _, cb := range CBlist {
		cb()
		// MOVQ	DX, main..autotmp_6(SP)
		// MOVQ	(AX)(DX*8), DX			<- DX = AX + DX*8, т.е. берем элемент слайса с индексом DX
		// MOVQ	(DX), AX				<- AX = *DX (из слайса взяли адрес и разыменовали его в AX)
		// DX указывает на структуру funcval, в которой есть адрес функции и контекст
		// Там первый элемент - адрес функции, далее контекст произвольной длины,
		// но контекст обрабатывает само замыкание, вызывающая сторона про него не знает
		// CALL	AX						<- вызов функции, в DX указатель на контекст
		// суть в том, что все элемены слайса вызываются одинаково, как замыкания, так и функции
		// MOVQ	main..autotmp_6(SP), DX <- инкремент текущего индекса итерации и другие команды для цикла
		// INCQ	DX
		// MOVQ	main..autotmp_4+16(SP), AX
		// MOVQ	main..autotmp_5+8(SP), CX
		// CMPQ	DX, CX
	}
}

//go:noinline
func vfunc() {
	println("vfunc")
}

func main() {
	AddCB(vfunc)
	// тут уже привычный код добавления элемента в слайс
	// MOVQ	main.CBlist+16(SB), CX
	// MOVQ	main.CBlist+8(SB), BX
	// INCQ	BX
	// XCHGL	AX, AX
	// MOVQ	main.CBlist(SB), AX
	// ... код проверки на переполнение слайса
	// LEAQ	main.vfunc·f(SB), CX  <- адрес контекста, а не функции, потому что main.vfunc·f, а не main.vfunc
	// MOVQ	CX, -8(AX)(BX*8)      <- запись в слайс адреса контекста

	var i int = 7
	AddCB(func() {
		// тут для анонимной функции создается контекст в динмаической памяти (runtime.newobject)
		// и также помещается в тот же слайс
		print("i = ", i, "\n")
		runtime.Caller(0)
	})

	CallCBs()
}
