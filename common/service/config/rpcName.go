package config

func (m *RPCName) UnmarshalYAML(
	unmarshal func(interface{}) error,
) error {

	var temp string
	if err := unmarshal(&temp); err != nil {
		return err
	}

	*m = "cadence-frontend"
	if "" != temp {
		*m = RPCName(temp)
	}

	return nil
}