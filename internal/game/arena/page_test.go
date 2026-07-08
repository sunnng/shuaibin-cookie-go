package arena

import "testing"

func TestNewPage(t *testing.T) {
	_ = NewPage(nil, nil, DefaultFeature())
}
