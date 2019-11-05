package generator

import (
	"strconv"
	"strings"

	"github.com/mapgoo-lab/atreus/tool/protobuf/pkg/generator"
	"github.com/mapgoo-lab/atreus/tool/protobuf/pkg/naming"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type ecode struct {
	generator.Base
	filesHandled int
}

// EcodeGenerator ecode generator.
func EcodeGenerator() *ecode {
	t := &ecode{}
	return t
}

// Generate ...
func (t *ecode) Generate(in *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
	t.Setup(in)

	// Showtime! Generate the response.
	resp := new(plugin.CodeGeneratorResponse)
	for _, f := range t.GenFiles {
		respFile := t.generateForFile(f)
		if respFile != nil {
			resp.File = append(resp.File, respFile)
		}
	}
	return resp
}

func (t *ecode) generateForFile(file *descriptor.FileDescriptorProto) *plugin.CodeGeneratorResponse_File {
	var enums []*descriptor.EnumDescriptorProto
	for _, enum := range file.EnumType {
		if strings.HasSuffix(*enum.Name, "ErrCode") {
			enums = append(enums, enum)
		}
	}
	if len(enums) == 0 {
		return nil
	}
	resp := new(plugin.CodeGeneratorResponse_File)
	t.generateFileHeader(file, t.GenPkgName)
	t.generateImports(file)
	for _, enum := range enums {
		t.generateEcode(file, enum)
	}

	resp.Name = proto.String(naming.GenFileName(file, ".ecode.go"))
	resp.Content = proto.String(t.FormattedOutput())
	t.Output.Reset()

	t.filesHandled++
	return resp
}

func (t *ecode) generateFileHeader(file *descriptor.FileDescriptorProto, pkgName string) {
	t.P("// Code generated by protoc-gen-ecode ", generator.Version, ", DO NOT EDIT.")
	t.P("// source: ", file.GetName())
	t.P()
	if t.filesHandled == 0 {
		comment, err := t.Reg.FileComments(file)
		if err == nil && comment.Leading != "" {
			// doc for the first file
			t.P("/*")
			t.P("Package ", t.GenPkgName, " is a generated ecode package.")
			t.P("This code was generated with atreus/tool/protobuf/protoc-gen-ecode ", generator.Version, ".")
			t.P()
			for _, line := range strings.Split(comment.Leading, "\n") {
				line = strings.TrimPrefix(line, " ")
				// ensure we don't escape from the block comment
				line = strings.Replace(line, "*/", "* /", -1)
				t.P(line)
			}
			t.P()
			t.P("It is generated from these files:")
			for _, f := range t.GenFiles {
				t.P("\t", f.GetName())
			}
			t.P("*/")
		}
	}
	t.P(`package `, pkgName)
	t.P()
}

func (t *ecode) generateImports(file *descriptor.FileDescriptorProto) {
	t.P(`import (`)
	t.P(`	"github.com/mapgoo-lab/atreus/pkg/ecode"`)
	t.P(`)`)
	t.P()
	t.P(`// to suppressed 'imported but not used warning'`)
	t.P(`var _ ecode.Codes`)
}

func (t *ecode) generateEcode(file *descriptor.FileDescriptorProto, enum *descriptor.EnumDescriptorProto) {
	t.P("// ", *enum.Name, " ecode")
	t.P("var (")

	for _, item := range enum.Value {
		if *item.Number == 0 {
			continue
		}
		// NOTE: eg: t.P("UserNotExist = New(-404) ")
		t.P(*item.Name, " = ", "ecode.New(", strconv.Itoa(int(*item.Number)), ")")
	}

	t.P(")")
}
