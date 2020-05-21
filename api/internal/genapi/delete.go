package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

type deleteInfo struct {
	baseType   string
	targetType string
	path       string
}

var deleteFuncs = map[string][]*deleteInfo{
	"scopes": {
		{
			"Organization",
			"Project",
			"projects/%s",
		},
	},
}

func writeDeleteFuncs() {
	for outPkg, funcs := range deleteFuncs {
		outFile := os.Getenv("GEN_BASEPATH") + fmt.Sprintf("/api/%s/delete.gen.go", outPkg)
		outBuf := bytes.NewBuffer([]byte(fmt.Sprintf(
			`// Code generated by "make api"; DO NOT EDIT.
package %s
`, outPkg)))
		for _, deleteInfo := range funcs {
			deleteFuncTemplate.Execute(outBuf, struct {
				BaseType        string
				TargetType      string
				LowerTargetType string
				Path            string
			}{
				BaseType:        deleteInfo.baseType,
				TargetType:      deleteInfo.targetType,
				LowerTargetType: strings.ToLower(deleteInfo.targetType),
				Path:            deleteInfo.path,
			})
		}
		if err := ioutil.WriteFile(outFile, outBuf.Bytes(), 0644); err != nil {
			fmt.Printf("error writing file %q: %v\n", outFile, err)
			os.Exit(1)
		}
	}
}

var deleteFuncTemplate = template.Must(template.New("").Parse(
	`
// Delete{{ .TargetType }} returns true iff the {{ .LowerTargetType }} existed when the delete attempt was made. 
func (s {{ .BaseType }}) Delete{{ .TargetType }}(ctx context.Context, {{ .LowerTargetType }} *{{ .TargetType }}) (bool, *api.Error, error) {
	if s.Client == nil {
		return false, nil, fmt.Errorf("nil client in Delete{{ .TargetType }} request")
	}
	if s.Id == "" {
		{{ if (eq .BaseType "Organization") }}
		// Assume the client has been configured with organization already and
		// move on
		{{ else if (eq .BaseType "Project") }}
		// Assume the client has been configured with project already and move
		// on
		{{ else }}
		return nil, nil, fmt.Errorf("missing {{ .BaseType }} ID in Delete{{ .TargetType }} request")
		{{ end }}
	} else {
		// If it's explicitly set here, override anything that might be in the
		// client
		{{ if (eq .BaseType "Organization") }}
		ctx = context.WithValue(ctx, "org", s.Id)
		{{ else if (eq .BaseType "Project") }}
		ctx = context.WithValue(ctx, "project", s.Id)
		{{ end }}
	}
	if {{ .LowerTargetType }}.Id == "" {
		return false, nil, fmt.Errorf("empty {{ .LowerTargetType }} ID field in Delete{{ .TargetType }} request")
	}

	req, err := s.Client.NewRequest(ctx, "DELETE", fmt.Sprintf("{{ .Path }}", {{ .LowerTargetType }}.Id), nil)
	if err != nil {
		return false, nil, fmt.Errorf("error creating Delete{{ .TargetType }} request: %w", err)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return false, nil, fmt.Errorf("error performing client request during Delete{{ .TargetType }} call: %w", err)
	}

	type deleteResponse struct {
		Existed bool
	}
	target := &deleteResponse{}

	apiErr, err := resp.Decode(target)
	if err != nil {
		return false, nil, fmt.Errorf("error decoding Delete{{ .TargetType }} repsonse: %w", err)
	}

	return target.Existed, apiErr, nil
}
`))