package main

// StringArray allows YAML properties to be parsed as either a single string or an array of strings
type StringArray []string

func (a *StringArray) UnmarshalYAML(unmarshal func(any) error) error {
	var multi []string
	err := unmarshal(&multi)
	if err != nil {
		var single string
		err := unmarshal(&single)
		if err != nil {
			return err
		}
		*a = []string{single}
	} else {
		*a = multi
	}
	return nil
}
