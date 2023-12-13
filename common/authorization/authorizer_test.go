package authorization

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidatePermissions(t *testing.T) {

	assert.Error(t, validatePermission(&JWTClaims{}, &Attributes{}, map[string]string{}), "no arguments should always fail")

}

func Test_validatePermission(t *testing.T) {

	attr := &Attributes{
		DomainName: "test-domain",
		TaskList:   nil,
		Permission: PermissionRead,
	}

	tests := []struct {
		name       string
		claims     *JWTClaims
		attributes *Attributes
		data       map[string]string
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name:       "no args should always fail",
			claims:     &JWTClaims{},
			attributes: &Attributes{},
			data:       map[string]string{},
			wantErr:    assert.Error,
		},
		{
			name:       "No claim defined should always fail",
			claims:     &JWTClaims{},
			attributes: attr,
			data:       map[string]string{},
			wantErr:    assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, validatePermission(tt.claims, tt.attributes, tt.data))
		})
	}
}
