package callcache

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestDispatcher_Do(t *testing.T) {
	type value struct {
		data string
	}
	testDispatcherDo[*value](t, "value", []testDispatcherDoTest[*value]{
		{
			name:       "ok",
			key:        "key",
			fn:         func() (*value, error) { return &value{data: "result"}, nil },
			want:       &value{data: "result"},
			wantUpdate: true,
		},
		{
			name:    "err",
			key:     "key",
			fn:      func() (*value, error) { return nil, errors.New("call error") },
			want:    nil,
			wantErr: true,
		},
		{
			name:       "nil",
			key:        "key",
			fn:         func() (*value, error) { return nil, nil },
			want:       nil,
			wantUpdate: true,
		},
	})

	testDispatcherDo[context.Context](t, "interface", []testDispatcherDoTest[context.Context]{
		{
			name:       "ok",
			key:        "key",
			fn:         func() (context.Context, error) { return context.TODO(), nil },
			want:       context.TODO(),
			wantUpdate: true,
		},
		{
			name:    "err",
			key:     "key",
			fn:      func() (context.Context, error) { return nil, errors.New("call error") },
			want:    nil,
			wantErr: true,
		},
		{
			name:       "nil",
			key:        "key",
			fn:         func() (context.Context, error) { return nil, nil },
			want:       nil,
			wantUpdate: true,
		},
	})
}

type testDispatcherDoTest[V any] struct {
	name       string
	key        string
	fn         func() (V, error)
	want       V
	wantErr    bool
	wantUpdate bool
}

func testDispatcherDo[V any](t *testing.T, prefix string, tests []testDispatcherDoTest[V]) {
	t.Helper()
	for _, tt := range tests {
		t.Run(prefix+"/"+tt.name, func(t *testing.T) {
			d := NewDispatcher[V](0, 0)

			got, err := d.Do(tt.key, tt.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Do() = %v, want %v", got, tt.want)
			}

			c, loaded := d.calls.Load(tt.key)
			if !loaded {
				t.Fatal("expected entry in calls map")
			}
			call := c.(*call[V])
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
		data map[string]string
	}{
		{
			name: "ok",
			args: args{key: "key"},
			data: map[string]string{"key": "result"},
		},
		{
			name: "missing",
			args: args{key: "key"},
			data: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDispatcher[string](0, 0)
			for k, v := range tt.data {
				d.calls.Store(k, &call[string]{result: v})
			}

			d.Remove(tt.args.key)
			if _, loaded := d.calls.Load(tt.args.key); loaded {
				t.Error("calls map should not contain removed key")
			}
		})
	}
}
