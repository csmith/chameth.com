---
title: Alternative YAML representations
group: Go
---

Sometimes YAML formats accept various data types, such as flipping between a string for simple config and a more complex object for advanced options. To handle this in Go:

```go
func (m *MyStruct) UnmarshalYAML(unmarshal func(interface{}) error) error {
    // Try to unmarshal as a plain string first
	str := ""
	if err := unmarshal(&str); err == nil {
		m.SimpleOption = str
		m.OtherOption = "default"
		return nil
	}

	// Cast to a type that doesn't implement yaml.Unmarshaler and carry on.
	type bare MyStruct
	return unmarshal((*bare)(m))
}
```