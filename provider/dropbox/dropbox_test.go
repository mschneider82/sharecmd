package dropbox

import (
	"os"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/mschneider82/sharecmd/config"
)

func TestDropboxProvider_Upload(t *testing.T) {
	cfg, err := config.lookupConfig("config.json")
	if err != nil {
		t.Fatalf("fail: %s", err.Error())
	}

	f, _ := os.Open("/etc/hosts")
	type fields struct {
		Config dropbox.Config
		token  string
	}
	type args struct {
		file *os.File
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			fields: fields{
				Config: dropbox.Config{
					Token:    cfg.ProviderSettings["token"],
					LogLevel: dropbox.LogOff,
				},
			},
			args: args{
				file: f,
				path: "/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &DropboxProvider{
				Config: tt.fields.Config,
				token:  tt.fields.token,
			}
			if _, err := c.Upload(tt.args.file, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("DropboxProvider.Upload() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func TestDropboxProvider_GetLink(t *testing.T) {
	cfg, err := lookupConfig()
	if err != nil {
		t.Fatalf("fail: %s", err.Error())
	}

	type fields struct {
		Config dropbox.Config
		token  string
	}
	type args struct {
		filepath string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "test1",
			fields: fields{
				Config: dropbox.Config{
					Token:    cfg.ProviderSettings["token"],
					LogLevel: dropbox.LogOff,
				},
			},
			args: args{
				filepath: "/hosts",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &DropboxProvider{
				Config: tt.fields.Config,
				token:  tt.fields.token,
			}
			link, err := c.GetLink(tt.args.filepath)
			if err != nil {
				t.Error(err)
			} else {
				t.Logf("link: %s", link)
			}
		})
	}
}
