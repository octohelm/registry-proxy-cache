package kubernetes

import (
	"context"
	"testing"
)

func TestListClusterImages(t *testing.T) {
	_, err := GetClusterContainerImages(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	//spew.Dump(images)
}
