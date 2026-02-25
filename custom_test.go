package toon

import (
	"strconv"
	"testing"
	"time"
)

// Custom Status type
type Status int

const (
	StatusPending Status = iota
	StatusActive
	StatusInactive
)

// MarshalTOON implements Marshaler
func (s Status) MarshalTOON() ([]byte, error) {
	switch s {
	case StatusPending:
		return []byte("P"), nil
	case StatusActive:
		return []byte("A"), nil
	case StatusInactive:
		return []byte("I"), nil
	}
	return []byte("?"), nil
}

// UnmarshalTOON implements Unmarshaler
func (s *Status) UnmarshalTOON(data []byte) error {
	if len(data) != 1 {
		return ErrMalformedTOON
	}
	switch data[0] {
	case 'P':
		*s = StatusPending
	case 'A':
		*s = StatusActive
	case 'I':
		*s = StatusInactive
	default:
		return ErrMalformedTOON
	}
	return nil
}

// Time wrapper for Unix timestamp
type UnixTime time.Time

func (t UnixTime) MarshalTOON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

func (t *UnixTime) UnmarshalTOON(data []byte) error {
	unix, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*t = UnixTime(time.Unix(unix, 0))
	return nil
}

func TestCustomMarshaler(t *testing.T) {
	type Task struct {
		ID     int
		Status Status
	}

	task := &Task{ID: 42, Status: StatusActive}
	data, err := Marshal(task)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	expected := "task{id,status}:42,A"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}

	// Unmarshal
	var decoded Task
	err = Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.ID != 42 || decoded.Status != StatusActive {
		t.Errorf("expected {42 A}, got {%d %v}", decoded.ID, decoded.Status)
	}
}

func TestUnixTimeMarshaler(t *testing.T) {
	type Event struct {
		Name string
		Time UnixTime
	}

	event := &Event{
		Name: "launch",
		Time: UnixTime(time.Unix(1609459200, 0)), // 2021-01-01 00:00:00 UTC
	}

	data, err := Marshal(event)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	expected := "event{name,time}:launch,1609459200"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}

	// Unmarshal
	var decoded Event
	err = Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Name != "launch" {
		t.Errorf("name: expected launch, got %s", decoded.Name)
	}
	if time.Time(decoded.Time).Unix() != 1609459200 {
		t.Errorf("time: expected 1609459200, got %d", time.Time(decoded.Time).Unix())
	}
}

func BenchmarkCustomMarshaler(b *testing.B) {
	type Task struct {
		ID     int
		Status Status
	}
	task := &Task{ID: 42, Status: StatusActive}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(task)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCustomUnmarshaler(b *testing.B) {
	data := []byte("task{id,status}:42,A")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		type Task struct {
			ID     int
			Status Status
		}
		var task Task
		err := Unmarshal(data, &task)
		if err != nil {
			b.Fatal(err)
		}
	}
}
