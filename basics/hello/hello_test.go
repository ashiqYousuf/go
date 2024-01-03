package hello

import "testing"

func TestSayHello(t *testing.T) {
	subtests := []struct {
		items  []string
		result string
	}{
		{
			result: "Hello, world!",
		},
		{
			result: "Hello, ashiq!",
			items:  []string{"ashiq"},
		},
		{
			result: "Hello, ashiq, hussain!",
			items:  []string{"ashiq", "hussain"},
		},
	}

	for _, st := range subtests {
		if s := Say(st.items); s != st.result {
			t.Errorf("wanted: %s [%v], got %s", st.items, st.result, s)
		}
	}
}
