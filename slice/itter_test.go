package slice

import (
	"reflect"
	"testing"
)

func TestItter_Pop(t *testing.T) {
	type fields struct {
		ready  bool
		offset int
		data   []int
	}
	type args struct {
		n int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       []int
		wantOffset int
	}{
		{
			"pop len 1",
			fields{
				ready:  true,
				offset: 4,
				data:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			},
			args{
				n: 1,
			},
			[]int{4},
			5,
		},
		{
			"pop len 5",
			fields{
				ready:  true,
				offset: 2,
				data:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			},
			args{
				n: 5,
			},
			[]int{2, 3, 4, 5, 6},
			7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := Itter[int]{
				offset: tt.fields.offset,
				data:   tt.fields.data,
			}
			if got := i.Pop(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Itter.Pop() = %v, want %v", got, tt.want)
			}
			if i.offset != tt.wantOffset {
				t.Errorf("Itter.Pop() = %v, want %v", i.offset, tt.wantOffset)
			}
		})
	}
}

func TestItter_Peek(t *testing.T) {
	type fields struct {
		ready  bool
		offset int
		data   []int
	}
	type args struct {
		n int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       []int
		wantOffset int
	}{
		{
			"peek 1",
			fields{
				offset: 4,
				data:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			},
			args{
				n: 1,
			},
			[]int{4},
			4,
		},
		{
			"peek 5",
			fields{
				offset: 2,
				data:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			},
			args{
				n: 5,
			},
			[]int{2, 3, 4, 5, 6},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := Itter[int]{
				offset: tt.fields.offset,
				data:   tt.fields.data,
			}
			if got := i.Peek(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Itter.Pop() = %v, want %v", got, tt.want)
			}
			if i.offset != tt.wantOffset {
				t.Errorf("Itter.Pop() = %v, want %v", i.offset, tt.wantOffset)
			}
		})
	}
}
