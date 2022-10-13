package config

import (
	"testing"
)

func TestInitConfig(t *testing.T) {
	type args struct {
		filePath []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "TestInitConfig",
			args:    args{filePath: []string{"../../conf/config.yaml"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InitFileConfig(tt.args.filePath...); (err != nil) != tt.wantErr {
				t.Errorf("InitFileConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Logf("%#v", globalConfig)
		})
	}
}

func TestInitFromEnvConfig(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{ // Modify run configuration & set environment variables
			name:    "TestInitFromEnvConfig",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InitFromEnvConfig(); (err != nil) != tt.wantErr {
				t.Errorf("InitFromEnvConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Logf("globalConfig:%+v", globalConfig)
		})
	}
}
