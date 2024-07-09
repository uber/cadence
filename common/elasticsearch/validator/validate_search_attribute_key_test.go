package validator

import (
	"testing"
)

func TestValidateSearchAttributeKey(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "pure character",
			args: args{name: "CustomStringField"},
		},
		{
			name: "alphanumeric",
			args: args{name: "CustomStringField1"},
		},
		{
			name:    "start with number",
			args:    args{name: "1CustomStringField"},
			wantErr: true,
		},
		{
			name:    "contain special character",
			args:    args{name: "CustomStringField!"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateSearchAttributeKey(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("ValidateSearchAttributeKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeSearchAttributeKey(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "pure character",
			args: args{name: "CustomStringField"},
			want: "CustomStringField",
		},
		{
			name: "alphanumeric",
			args: args{name: "CustomStringField1"},
			want: "CustomStringField1",
		},
		{
			name:    "start with number",
			args:    args{name: "1CustomStringField"},
			want:    "CustomStringField",
		},
		{
			name:   "contain special character",
			args:   args{name: "CustomStringField!"},
			want:   "CustomStringField",
		},
		{
			name:   "contain special character in the middle",
			args:   args{name: "CustomString-Field"},
			want:   "CustomString_Field",
		},
		{
			name:   "all numbers",
			args:   args{name: "1234567890"},
			want:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeSearchAttributeKey(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeSearchAttributeKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizeSearchAttributeKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
