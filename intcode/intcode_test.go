package intcode

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func Test_decodeOp(t *testing.T) {
	type args struct {
		op int
	}
	tests := []struct {
		name string
		args args
		want OpCode
	}{
		{"test1", args{1002}, OpCode{op: 2, parmModes: [3]int{0, 1, 0}}},
		{"test2", args{11199}, OpCode{op: 99, parmModes: [3]int{1, 1, 1}}},
		{"test3", args{42}, OpCode{op: 42, parmModes: [3]int{}}},
		{"test4", args{10011}, OpCode{op: 11, parmModes: [3]int{0, 0, 1}}},
		{"test5", args{102}, OpCode{op: 2, parmModes: [3]int{1, 0, 0}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeOp(tt.args.op); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVM_ExecPgm(t *testing.T) {
	type fields struct {
		ID     int
		Pgm    *Program
		Input  chan int
		Output chan int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"test1", fields{Pgm: Compile(`109,1,204,-1,1001,100,1,100,1008,100,16,101,1006,101,0,99`), Output: make(chan int)}, `109,1,204,-1,1001,100,1,100,1008,100,16,101,1006,101,0,99`},
		{"test2", fields{Pgm: Compile(`1102,34915192,34915192,7,4,7,99,0`), Output: make(chan int)}, `1219070632396864`},
		{"test3", fields{Pgm: Compile(`104,1125899906842624,99`), Output: make(chan int)}, `1125899906842624`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				err error
				wg sync.WaitGroup
			)
			vm := &VM{
				ID:     tt.fields.ID,
				Pgm:    tt.fields.Pgm,
				Input:  tt.fields.Input,
				Output: tt.fields.Output,
			}
			wg.Add(1)
			output := make([]string, 0)
			done := make(chan struct{})
			go func() {
				for o := range tt.fields.Output {
					output = append(output, strconv.Itoa(o))
				}
				close(done)
			}()
			go func() {
				err = vm.ExecPgm()
				wg.Done()
			}()
			wg.Wait()
			close(vm.Output)
			<- done
			if err != nil {
				t.Errorf("ExecPgm() error = %v", err)
			} else {
				got := strings.Join(output, ",")
				if got != tt.want {
					t.Errorf("ExecPgm() error, want '%s' got '%s'", tt.want, got)

				}
			}
		})
	}
}
