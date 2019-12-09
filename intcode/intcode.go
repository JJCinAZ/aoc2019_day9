package intcode

import (
	"fmt"
	"strconv"
	"strings"
)

type OpCode struct {
	op        int
	parmModes [3]int
}

type VM struct {
	ID     int
	Pgm    *Program
	Input  chan int
	Output chan int
}

type Program struct {
	code  []int
	ip    int
	base  int
	debug bool
}

func NewVM(id int, pgm *Program, in, out chan int) *VM {
	vm := new(VM)
	vm.ID = id
	vm.Pgm = pgm.Copy()
	vm.Input = in
	vm.Output = out
	return vm
}

func Compile(input string) *Program {
	pgm := new(Program)
	a := strings.Split(input, ",")
	pgm.code = make([]int, len(a)+16384)
	for i := range a {
		pgm.code[i], _ = strconv.Atoi(a[i])
	}
	return pgm
}

func (p *Program) Debug(b bool) {
	p.debug = b
}

func (p *Program) Copy() *Program {
	pgm := new(Program)
	pgm.code = make([]int, len(p.code))
	copy(pgm.code, p.code)
	return pgm
}

func decodeOp(op int) OpCode {
	result := OpCode{}
	result.parmModes[2] = op / 10000
	op = op % 10000
	result.parmModes[1] = op / 1000
	op = op % 1000
	result.parmModes[0] = op / 100
	result.op = op % 100
	return result
}

func (vm *VM) ExecPgm() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	p := vm.Pgm
	p.ip = 0
PGMLOOP:
	for {
		opcode := decodeOp(p.code[p.ip])
		if p.debug {
			fmt.Println(opcode)
		}
		switch opcode.op {
		case 99:
			if p.debug {
				fmt.Printf("%02d: HALT\n", vm.ID)
			}
			//vm.Output <- 999
			break PGMLOOP
		case 1: // Addition
			v1, v2 := p.getParamsValues(opcode)
			p.setParamValue(opcode.parmModes[2], 3, v1 + v2)
			p.ip += 4
		case 2: // Multiplication
			v1, v2 := p.getParamsValues(opcode)
			p.setParamValue(opcode.parmModes[2], 3, v1 * v2)
			p.ip += 4
		case 3: // Input
			var b int
			b = <-vm.Input
			p.setParamValue(opcode.parmModes[0], 1, b)
			if p.debug {
				fmt.Printf("%02d: INPUT:%d\n", vm.ID, b)
			}
			p.ip += 2
		case 4: // Output
			v1 := p.getParamValue(opcode.parmModes[0], 1)
			vm.Output <- v1
			if p.debug {
				fmt.Printf("%02d: OUTPUT:%d\n", vm.ID, v1)
			}
			p.ip += 2
		case 5: // Jump-if-true
			v1, v2 := p.getParamsValues(opcode)
			if v1 != 0 {
				p.ip = v2
			} else {
				p.ip += 3
			}
		case 6: // Jump-if-false
			v1, v2 := p.getParamsValues(opcode)
			if v1 == 0 {
				p.ip = v2
			} else {
				p.ip += 3
			}
		case 7: // Less-than
			v1, v2 := p.getParamsValues(opcode)
			if v1 < v2 {
				p.setParamValue(opcode.parmModes[2], 3, 1)
			} else {
				p.setParamValue(opcode.parmModes[2], 3, 0)
			}
			p.ip += 4
		case 8: // Equals
			v1, v2 := p.getParamsValues(opcode)
			if v1 == v2 {
				p.setParamValue(opcode.parmModes[2], 3, 1)
			} else {
				p.setParamValue(opcode.parmModes[2], 3, 0)
			}
			p.ip += 4
		case 9: // Adjust relative base
			v1 := p.getParamValue(opcode.parmModes[0], 1)
			p.base += v1
			p.ip += 2
		default:
			panic(fmt.Errorf("illegal opcode at offset %d", p.ip))
		}
	}
	return nil
}

func (p *Program) getParamsValues(opcode OpCode) (int, int) {
	return p.getParamValue(opcode.parmModes[0], 1), p.getParamValue(opcode.parmModes[1], 2)
}

func (p *Program) getParamValue(parmMode int, ipOffset int) int {
	switch parmMode {
	case 0: // Position Mode
		return p.code[p.code[p.ip+ipOffset]]
	case 1: // Immediate Mode
		return p.code[p.ip+ipOffset]
	case 2: // Relative Mode
		return p.code[p.code[p.ip+ipOffset]+p.base]
	}
	panic(fmt.Errorf("illegal parameter mode at offset %d", p.ip))
	return 0
}

func (p *Program) setParamValue(parmMode int, ipOffset int, value int) {
	switch parmMode {
	case 0: // Position Mode
		p.code[p.code[p.ip+ipOffset]] = value
	case 1: // Immediate Mode
		p.code[p.ip+ipOffset] = value
	case 2: // Relative Mode
		p.code[p.code[p.ip+ipOffset]+p.base] = value
	default:
		panic(fmt.Errorf("illegal parameter mode at offset %d", p.ip))
	}
}
