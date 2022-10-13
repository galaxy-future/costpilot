package tools

import "testing"

func TestGetCurrentDirectory(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "TestGetCurrentDirectory",
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCurrentDirectory()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetCurrentDirectory() got = %v, want %v", got, tt.want)
			}
		})
	}
}
