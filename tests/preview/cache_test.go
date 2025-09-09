package preview

import (
	"os"
	"testing"

	"github.com/bravo1goingdark/mailgrid/utils/preview"
)

func TestLoadTemplateCaching(t *testing.T) {
	tmplContent := "<p>{{.Email}}</p>"
	tmp, err := os.CreateTemp(t.TempDir(), "tmpl*.html")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err := os.Remove(tmp.Name())
		if err != nil {
			return
		}
	})
	if _, err := tmp.WriteString(tmplContent); err != nil {
		t.Fatal(err)
	}
	err = tmp.Close()
	if err != nil {
		return
	}

	first, err := preview.LoadTemplate(tmp.Name())
	if err != nil {
		t.Fatalf("first load error: %v", err)
	}
	second, err := preview.LoadTemplate(tmp.Name())
	if err != nil {
		t.Fatalf("second load error: %v", err)
	}
	if first != second {
		t.Error("expected cached template to be reused")
	}
}
