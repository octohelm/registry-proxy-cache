package container

import "strings"

func ParseContainerImage(image string) *ContainerImage {
	parts := strings.SplitN(image, "/", 3)

	if len(parts) != 3 {
		return nil
	}

	ci := &ContainerImage{
		Hub:       parts[0],
		Namespace: parts[1],
	}

	nameAndTag := strings.Split(parts[2], ":")

	ci.Name = nameAndTag[0]

	if len(nameAndTag) == 1 {
		ci.Tag = "latest"
	} else {
		ci.Tag = nameAndTag[1]
	}

	return ci
}

type ContainerImage struct {
	Hub       string
	Namespace string
	Name      string
	Tag       string
}

func (ci *ContainerImage) Ref() string {
	return ci.Namespace + "/" + ci.Name + ":" + ci.Tag
}
