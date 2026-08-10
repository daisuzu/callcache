package callcache

import (
	"errors"
	"reflect"
	"testing"
)

func TestDispatcher_Do(t *testing.T) {
	type args struct {
		key string
		fn  func() (any, error)
	}
	tests := []struct {
		name       string
		args       args
		want       any
		wantErr    bool
		wantUpdate bool
	}{
		{
			name: "ok",
			args: args{
				key: "key",
				fn:  func() (any, error) { return "result", nil },
			},
			want:       "result",
			wantUpdate: true,
		},
		{
			name: "err",
			args: args{
				key: "key",
				fn:  func() (any, error) { return nil, errors.New("call error") },
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDispatcher(0, 0)

			got, err := d.Do(tt.args.key, tt.args.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Do() = %v, want %v", got, tt.want)
			}

			c, loaded := d.calls.Load(tt.args.key)
			if !loaded {
				t.Fatal("expected entry in calls map")
			}
			call := c.(*call)
			if !reflect.DeepEqual(call.result, tt.want) {
				t.Errorf("call.result = %v, want %v", call.result, tt.want)
			}
			if (call.lastUpdate != 0) != tt.wantUpdate {
				t.Errorf("call.lastUpdate(%d) != 0 is %v, want %v", call.lastUpdate, call.lastUpdate != 0, tt.wantUpdate)
			}
		})
	}
}

func TestDispatcher_Remove(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		data map[string]any
	}{
		{
			name: "ok",
			args: args{key: "key"},
			data: map[string]any{"key": "result"},
		},
		{
			name: "missing",
			args: args{key: "key"},
			data: map[string]any{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDispatcher(0, 0)
			for k, v := range tt.data {
				d.calls.Store(k, &call{result: v})
			}

			d.Remove(tt.args.key)
			if _, loaded := d.calls.Load(tt.args.key); loaded {
				t.Error("calls map should not contain removed key")
			}
		})
	}
}
