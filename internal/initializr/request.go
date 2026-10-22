package initializr

import (
	"net/url"
	"strings"
)

type GenerationRequest struct {
	Type, Language, BootVersion, BaseDir, GroupID, ArtifactID, Name, Description, PackageName, Packaging, JavaVersion string
	Dependencies                                                                                                      []string
}

func (c *Client) GenerationURL(request GenerationRequest) string {
	u := *c.BaseURL
	u.Path = strings.TrimRight(u.Path, "/") + "/starter.zip"
	values := url.Values{}
	set := func(key, value string) {
		if value != "" {
			values.Set(key, value)
		}
	}
	set("type", request.Type)
	set("language", request.Language)
	set("bootVersion", request.BootVersion)
	set("baseDir", request.BaseDir)
	set("groupId", request.GroupID)
	set("artifactId", request.ArtifactID)
	set("name", request.Name)
	set("description", request.Description)
	set("packageName", request.PackageName)
	set("packaging", request.Packaging)
	set("javaVersion", request.JavaVersion)
	if len(request.Dependencies) > 0 {
		values.Set("dependencies", strings.Join(request.Dependencies, ","))
	}
	u.RawQuery = values.Encode()
	return u.String()
}
