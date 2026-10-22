package validate

import "testing"

func TestCoordinates(t *testing.T) {
	for _, value := range []string{"com.example", "io.spring_create.demo"} {
		if err := GroupID(value); err != nil {
			t.Errorf("%q: %v", value, err)
		}
	}
	for _, value := range []string{"", "com..example", "com.example-name", "9com.example"} {
		if err := GroupID(value); err == nil {
			t.Errorf("%q should fail", value)
		}
	}
}

func TestSafeName(t *testing.T) {
	if err := SafeName("my-app"); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"", "../bad", "bad/name", "my project"} {
		if err := SafeName(value); err == nil {
			t.Errorf("%q should fail", value)
		}
	}
}
