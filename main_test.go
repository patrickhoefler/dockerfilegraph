package main

import "testing"

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "should work",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			main()
		})
	}
}
