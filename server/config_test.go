package server

import (
	"fmt"
	"testing"

	"github.com/flashbots/go-boost-utils/types"
	"github.com/stretchr/testify/require"
)

func _newRelayEntry(t *testing.T, relayURL string) RelayEntry {
	entry, err := NewRelayEntry(relayURL)
	require.NoError(t, err)

	return entry
}

func _newRelayEntries(t *testing.T, l, h int) []RelayEntry {
	var res []RelayEntry

	for i := l; i < h; i++ {
		pubKey := types.PublicKey{byte(i)}.String()
		newEntry := fmt.Sprintf("https://%s@%s%d%s", pubKey, "builder", i, "-relay-kiln.flashbots.net/")

		res = append(res, _newRelayEntry(t, newEntry))
	}

	return res
}

func TestCreateNewRawConfiguration(t *testing.T) {
	testCases := []struct {
		name     string
		filename string

		expectedError                bool
		expectedRawConfigurationFile *rawConfigurationFile
	}{
		{
			name:                         "It detects non-existing file",
			filename:                     "deadbeef",
			expectedError:                true,
			expectedRawConfigurationFile: nil,
		},
		{
			name:                         "It detects invalid JSON",
			filename:                     "testdata/invalid_json.input",
			expectedError:                true,
			expectedRawConfigurationFile: nil,
		},
		{
			name:          "It creates a valid raw configuration from file",
			filename:      "testdata/valid_json.input",
			expectedError: false,
			expectedRawConfigurationFile: &rawConfigurationFile{
				BuilderRelaysGroups: make(map[string][]string),
				ProposerConfig:      make(map[string]rawConfiguration),
				DefaultConfig: rawConfiguration{
					ValidatorRegistration: struct {
						BuilderRelays []string `json:"builder_relays"`
						Enabled       bool     `json:"enabled"`
					}(struct {
						BuilderRelays []string
						Enabled       bool
					}{BuilderRelays: []string{}, Enabled: false}),
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			rcf, err := newRawConfigurationFile(tt.filename)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, rcf)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedRawConfigurationFile, rcf)
			}
		})
	}
}

func TestCreateNewConfigurationStorage(t *testing.T) {
	relay0 := fmt.Sprintf("https://%s@%s", types.PublicKey{0x00}.String(), "builder0-relay-kiln.flashbots.net/")
	relay1 := fmt.Sprintf("https://%s@%s", types.PublicKey{0x01}.String(), "builder1-relay-kiln.flashbots.net/")

	testCases := []struct {
		name    string
		rawConf *rawConfiguration
		groups  map[string][]string

		expectedError                bool
		expectedConfigurationStorage *ProposerConfig
	}{
		{
			name: "It detects invalid fee recipient",
			rawConf: &rawConfiguration{
				ValidatorRegistration: struct {
					BuilderRelays []string `json:"builder_relays"`
					Enabled       bool     `json:"enabled"`
				}{
					BuilderRelays: []string{
						relay0,
					},
				},
			},
			groups: map[string][]string{
				"groupA": {relay0},
			},
			expectedError:                true,
			expectedConfigurationStorage: nil,
		},
		{
			name: "It detects invalid relay0 entry in raw configuration",
			rawConf: &rawConfiguration{
				ValidatorRegistration: struct {
					BuilderRelays []string `json:"builder_relays"`
					Enabled       bool     `json:"enabled"`
				}{
					BuilderRelays: []string{
						"deadbeef",
					},
				},
			},
			groups: map[string][]string{
				"groupA": {relay0},
			},
			expectedError:                true,
			expectedConfigurationStorage: nil,
		},
		{
			name: "It detects empty group",
			rawConf: &rawConfiguration{
				ValidatorRegistration: struct {
					BuilderRelays []string `json:"builder_relays"`
					Enabled       bool     `json:"enabled"`
				}{
					BuilderRelays: []string{
						"groupA",
					},
				},
			},
			groups: map[string][]string{
				"groupA": {},
			},
			expectedError:                true,
			expectedConfigurationStorage: nil,
		},
		{
			name: "It detects invalid relay0 entry in group",
			rawConf: &rawConfiguration{
				ValidatorRegistration: struct {
					BuilderRelays []string `json:"builder_relays"`
					Enabled       bool     `json:"enabled"`
				}{
					BuilderRelays: []string{
						"groupA",
					},
				},
			},
			groups: map[string][]string{
				"groupA": {
					"deadbeef",
				},
			},
			expectedError:                true,
			expectedConfigurationStorage: nil,
		},
		{
			name: "It detects empty relay array in proposer configuration",
			rawConf: &rawConfiguration{
				ValidatorRegistration: struct {
					BuilderRelays []string `json:"builder_relays"`
					Enabled       bool     `json:"enabled"`
				}{
					BuilderRelays: []string{},
				},
			},
			groups: map[string][]string{
				"groupA": {
					"deadbeef",
				},
			},
			expectedError:                true,
			expectedConfigurationStorage: nil,
		},
		{
			name: "It creates valid configuration storage from group only",
			rawConf: &rawConfiguration{
				ValidatorRegistration: struct {
					BuilderRelays []string `json:"builder_relays"`
					Enabled       bool     `json:"enabled"`
				}{
					BuilderRelays: []string{
						"groupA",
					},
				},
			},
			groups: map[string][]string{
				"groupA": {
					relay0,
				},
			},
			expectedError: false,
			expectedConfigurationStorage: &ProposerConfig{
				Enabled: false,
				Relays:  _newRelayEntries(t, 0, 1),
			},
		},
		{
			name: "It creates valid configuration storage from raw relay0 entries only",
			rawConf: &rawConfiguration{
				ValidatorRegistration: struct {
					BuilderRelays []string `json:"builder_relays"`
					Enabled       bool     `json:"enabled"`
				}{
					BuilderRelays: []string{
						relay0,
					},
				},
			},
			groups:        map[string][]string{},
			expectedError: false,
			expectedConfigurationStorage: &ProposerConfig{
				Enabled: false,
				Relays:  _newRelayEntries(t, 0, 1),
			},
		},
		{
			name: "It creates valid configuration storage from both raw relay0 entries and groups",
			rawConf: &rawConfiguration{
				ValidatorRegistration: struct {
					BuilderRelays []string `json:"builder_relays"`
					Enabled       bool     `json:"enabled"`
				}{
					BuilderRelays: []string{
						"groupA",
						relay1,
					},
				},
			},
			groups: map[string][]string{
				"groupA": {
					relay0,
				},
			},
			expectedError: false,
			expectedConfigurationStorage: &ProposerConfig{
				Enabled: false,
				Relays:  _newRelayEntries(t, 0, 2),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			configurationStorage, err := newConfigurationStorage(tt.rawConf, tt.groups)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, configurationStorage)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedConfigurationStorage, configurationStorage)
			}
		})
	}
}

func TestCreateNewProposerConfigurationStorage(t *testing.T) {
	testCases := []struct {
		name     string
		filename string

		expectedError                        bool
		expectedProposerConfigurationStorage *ProposerConfigurationStorage
	}{
		{
			name:          "It creates a valid raw configuration from file",
			filename:      "testdata/valid_config.json",
			expectedError: false,
			expectedProposerConfigurationStorage: &ProposerConfigurationStorage{
				proposerConfigurations: map[types.PublicKey]*ProposerConfig{
					_HexToPubkey("0xa057816155ad77931185101128655c0191bd0214c201ca48ed887f6c4c6adf334070efcd75140eada5ac83a92506dd7a"): {
						Enabled: true,
						Relays:  _newRelayEntries(t, 2, 6),
					},
				},
				defaultConfiguration: &ProposerConfig{
					Enabled: false,
					Relays:  _newRelayEntries(t, 6, 7),
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewProposerConfigurationStorage(tt.filename)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, storage)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedProposerConfigurationStorage, storage)
			}
		})
	}
}

func TestGetProposerConfiguration(t *testing.T) {
	proposerPubKey := _HexToPubkey("0xa057816155ad77931185101128655c0191bd0214c201ca48ed887f6c4c6adf334070efcd75140eada5ac83a92506dd7a")

	testCases := []struct {
		name    string
		storage ProposerConfigurationStorage

		expectedConfiguration *ProposerConfig
	}{
		{
			name: "It gets specific configuration",
			storage: ProposerConfigurationStorage{
				proposerConfigurations: map[types.PublicKey]*ProposerConfig{
					proposerPubKey: {
						Enabled: true,
						Relays:  _newRelayEntries(t, 0, 1),
					},
				},
			},
			expectedConfiguration: &ProposerConfig{
				Enabled: true,
				Relays:  _newRelayEntries(t, 0, 1),
			},
		},
		{
			name: "It gets default configuration",
			storage: ProposerConfigurationStorage{
				defaultConfiguration: &ProposerConfig{
					Enabled: true,
					Relays:  _newRelayEntries(t, 0, 2),
				},
			},
			expectedConfiguration: &ProposerConfig{
				Enabled: true,
				Relays:  _newRelayEntries(t, 0, 2),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			configurationStorage := tt.storage.GetProposerConfiguration(proposerPubKey)

			require.NotNil(t, configurationStorage)
			require.Equal(t, tt.expectedConfiguration, configurationStorage)
		})
	}
}
