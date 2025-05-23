package wardley

import (
	_ "embed"
	"net/http"
	"strings"
	"testing"
)

//go:embed assets/map.owm
var validMap string

//go:embed assets/invalid-map.owm
var invalidMap string

func TestRenderEngineRender(t *testing.T) {
	cases := []struct {
		content/*, result */ string
		errHasPrefix string
	}{
		{content: validMap},
		{content: invalidMap,
			errHasPrefix: "syntax error"},
		{content: "",
			errHasPrefix: "no map provided"},
	}
	re1, err := NewRenderEngine("0")
	if err != nil {
		t.Errorf("NewRenderEngine() error = %v", err)
	}

	defer re1.Cancel()
	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			got, err := re1.Render(tt.content)
			t.Logf("from  %s got %s, error %v", tt.content, got, err)
			if err != nil {
				if strings.HasPrefix(err.Error(), tt.errHasPrefix) {
					// expected exception
					return
				}
				t.Errorf("Render() error = %v", err)
			}
			if !strings.HasPrefix(string(got), "<svg") {
				t.Errorf("Render() got an invalid svg = %s, err = %v", got, err)
			}

			resultInBytes, box, err := re1.RenderAsPng(tt.content)
			if err != nil {
				if !strings.HasPrefix(err.Error(), tt.errHasPrefix) {
					t.Errorf("Render() error = %v", err)
					return
				}
			}
			if box == nil {
				t.Errorf("RenderAsPng() returned an empty box")
			} else if box.Width < 1 || box.Height < 1 {
				t.Errorf("RenderAsPng() got empty image = w:%d, h:%d)", box.Width, box.Height)
			}
			contentType := http.DetectContentType(resultInBytes)
			if contentType != "image/png" {
				t.Errorf("RenderAsPng() return an '%s' rather than 'image/png'", contentType)
			}
		})
	}
}

func BenchmarkRenderEngineRender(b *testing.B) {
	re1, _ := NewRenderEngine("0")
	for i := 0; i < b.N; i++ {
		_, _ = re1.Render(validMap)
	}
	re1.Cancel()
}
