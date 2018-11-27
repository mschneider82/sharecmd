package main

import (
	"fmt"
	"testing"
)

func TestGoogleDriveProvider_GetLink(t *testing.T) {
	cfg, err := lookupConfig()
	if err != nil {
		panic(fmt.Sprintf("lookupConfig: %v \n", err))
	}
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		fields  *GoogleDriveProvider
		args    args
		want    string
		wantErr bool
	}{
		{
			name:   "test1",
			fields: NewGoogleDriveProvider(cfg.ProviderSettings["googletoken"]),
			args: args{
				filepath: "/etc/hosts",
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields
			got, err := c.GetLink(tt.args.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GoogleDriveProvider.GetLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GoogleDriveProvider.GetLink() = %v, want %v", got, tt.want)
			}
		})
	}
}
