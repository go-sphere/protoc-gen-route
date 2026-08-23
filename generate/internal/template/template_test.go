package template

import (
	"strings"
	"testing"
)

func TestExecuteQuotesExtra(t *testing.T) {
	out, err := (&ServiceDesc{
		OptionsKey:  "Bot",
		ServiceType: "Menu",
		ServiceName: "bot.v1.Menu",
		Methods: []*MethodDesc{{
			Name:         "Start",
			OriginalName: "Start",
			Request:      "StartRequest",
			Reply:        "StartResponse",
			Extra: map[string]string{
				`key "quoted"`: "line1\nline2",
			},
		}},
		Package: &PackageDesc{
			RequestType:      "Request",
			ResponseType:     "Response",
			ExtraDataType:    "Extra",
			NewExtraDataFunc: "NewExtra",
		},
	}).Execute()
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, `"key \"quoted\"": "line1\nline2"`) {
		t.Errorf("extra map not Go-quoted, got:\n%s", out)
	}
}
