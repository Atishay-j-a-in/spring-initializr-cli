package initializr

import "fmt"

type Metadata struct {
	Links        map[string]Link      `json:"_links"`
	Dependencies DependencyCapability `json:"dependencies"`
	Type         SelectCapability     `json:"type"`
	Packaging    SelectCapability     `json:"packaging"`
	JavaVersion  SelectCapability     `json:"javaVersion"`
	Language     SelectCapability     `json:"language"`
	BootVersion  SelectCapability     `json:"bootVersion"`
	GroupID      TextCapability       `json:"groupId"`
	ArtifactID   TextCapability       `json:"artifactId"`
	Name         TextCapability       `json:"name"`
	Description  TextCapability       `json:"description"`
	PackageName  TextCapability       `json:"packageName"`
}

type Link struct {
	Href      string `json:"href"`
	Templated bool   `json:"templated"`
}

type TextCapability struct {
	Type    string `json:"type"`
	Default string `json:"default"`
}

type SelectCapability struct {
	Type    string  `json:"type"`
	Default string  `json:"default"`
	Values  []Value `json:"values"`
}

type Value struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Action      string            `json:"action"`
	Tags        map[string]string `json:"tags"`
}

type DependencyCapability struct {
	Type   string            `json:"type"`
	Values []DependencyGroup `json:"values"`
}

type DependencyGroup struct {
	Name   string       `json:"name"`
	Values []Dependency `json:"values"`
}

type Dependency struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (m Metadata) Validate() error {
	for name, capability := range map[string]SelectCapability{
		"type": m.Type, "language": m.Language, "packaging": m.Packaging,
		"javaVersion": m.JavaVersion, "bootVersion": m.BootVersion,
	} {
		if capability.Default == "" || len(capability.Values) == 0 {
			return fmt.Errorf("metadata is missing usable %s values", name)
		}
		if !hasValue(capability, capability.Default) {
			return fmt.Errorf("metadata default %q is not available for %s", capability.Default, name)
		}
	}
	if m.GroupID.Default == "" || m.ArtifactID.Default == "" || m.PackageName.Default == "" {
		return fmt.Errorf("metadata is missing project coordinate defaults")
	}
	if len(m.Dependencies.Values) == 0 {
		return fmt.Errorf("metadata is missing dependency groups")
	}
	return nil
}

func hasValue(capability SelectCapability, id string) bool {
	for _, value := range capability.Values {
		if value.ID == id {
			return true
		}
	}
	return false
}

func (c SelectCapability) Find(id string) (Value, bool) {
	for _, value := range c.Values {
		if value.ID == id {
			return value, true
		}
	}
	return Value{}, false
}
