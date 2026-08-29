package toml

import "testing"

type benchmarkConfig struct {
	Mode   string   `toml:"mode"`
	Output string   `toml:"output"`
	Flags  []string `toml:"flags"`
}

func TestUnmarshalMalformedReturnsError(t *testing.T) {
	var cfg benchmarkConfig
	if err := Unmarshal([]byte("mode = ["), &cfg); err == nil {
		t.Fatal("expected malformed TOML error")
	}
}

func TestUnmarshalEmptyInput(t *testing.T) {
	var cfg benchmarkConfig
	if err := Unmarshal(nil, &cfg); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	data := []byte("mode = \"raw\"\noutput = \"app\"\nflags = [\"-O2\", \"-pipe\"]\n")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg benchmarkConfig
		if err := Unmarshal(data, &cfg); err != nil {
			b.Fatal(err)
		}
	}
}
