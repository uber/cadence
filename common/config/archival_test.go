package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common"
)

func TestValidEnabledHistoryArchivalConfig(t *testing.T) {
	archival := Archival{
		History: HistoryArchival{
			Status: common.ArchivalEnabled,
			Provider: &HistoryArchiverProvider{
				Filestore: &FilestoreArchiver{
					FileMode: "044",
				},
			},
		},
	}
	err := archival.Validate(&ArchivalDomainDefaults{
		History: HistoryArchivalDomainDefaults{
			URI: "/var/tmp",
		},
	})
	require.NoError(t, err)
}

func TestInvalidHEnabledHistoryArchivalConfig(t *testing.T) {
	archival := Archival{
		History: HistoryArchival{
			Status: common.ArchivalEnabled,
		},
	}
	err := archival.Validate(&ArchivalDomainDefaults{})
	require.Error(t, err)
}

func TestValidDisabledHistoryArchivalConfig(t *testing.T) {
	archival := Archival{
		History: HistoryArchival{
			Provider: &HistoryArchiverProvider{
				Filestore: &FilestoreArchiver{
					FileMode: "044",
				},
			},
		},
	}
	err := archival.Validate(&ArchivalDomainDefaults{})
	require.NoError(t, err)
}

func TestValidEmptyHistoryArchivalConfig(t *testing.T) {
	archival := Archival{
		History: HistoryArchival{
		},
	}
	err := archival.Validate(&ArchivalDomainDefaults{})
	require.NoError(t, err)
}