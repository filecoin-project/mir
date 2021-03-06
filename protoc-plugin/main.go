package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/filecoin-project/mir/pkg/pb/mir"
)

// generateGenericFriendlyEnumsForEventTypes generates public interfaces of the form "[Msg]_[Oneof]" and
// "[Msg]_[Oneof]Wrapper", for each oneof marked with "option (mir.event_type) = true", where [Msg] is the name of the
// message and [Oneof] is the name of the oneof field. For example, for a oneof field called "Type" in a message called
// "Event", the generated interfaces will be "Event_Type" and "Event_TypeWrapper".
// "Event_Type" represents all types assignable to the field Type of the message of type Event.
// "Event_TypeWrapper[T]" represents all types assignable to the field Type of the message of type Event such that the
// underlying type of the oneof wrapper is T.
// Additionally, for all oneof wrappers generated by protoc for oneofs annotated with "option (mir.event_type) = true",
// method "Unwrap()" that returns the underlying object is added.
func generateGenericFriendlyEnumsForEventTypes(plugin *protogen.Plugin, file *protogen.File) error {
	var g *protogen.GeneratedFile

	for _, msg := range file.Messages {
		for _, oneof := range msg.Oneofs {
			oneofOptions := oneof.Desc.Options().(*descriptorpb.OneofOptions)
			markedAsEventType := proto.GetExtension(oneofOptions, mir.E_EventType).(bool)
			if !markedAsEventType {
				continue
			}

			if g == nil {
				filename := fmt.Sprintf("%s.pb.mir.go", file.GeneratedFilenamePrefix)
				g = plugin.NewGeneratedFile(filename, file.GoImportPath)
				g.P("package ", file.GoPackageName)
				g.P()
			}

			interfaceName := g.QualifiedGoIdent(oneof.GoIdent)
			g.P("type ", interfaceName, " = ", "is", interfaceName)
			g.P()

			g.P("type ", interfaceName, "Wrapper[Ev any] interface {")
			g.P("\t", interfaceName)
			g.P("\t", "Unwrap() *Ev")
			g.P("}")
			g.P()

			for _, field := range oneof.Fields {
				wrapperTypeName := g.QualifiedGoIdent(field.GoIdent)
				_ = wrapperTypeName
				if field.Desc.Kind() != protoreflect.MessageKind {
					return fmt.Errorf("oneof field \"%v\" annotated with \"option (mir.event_type) = true\" "+
						"is supposed to have only Message entries (no primitive types, no nested oneofs). "+
						"Field \"%v\", expected kind: %v, actual kind: %v",
						oneof.Desc.Name(), field.Desc.Name(), protoreflect.MessageKind, field.Desc.Kind())
				}
				typeName := g.QualifiedGoIdent(field.Message.GoIdent)

				g.P("func (p *", wrapperTypeName, ") Unwrap() *", typeName, " {")
				g.P("\treturn p.", field.GoName)
				g.P("}")
				g.P()
			}
		}
	}

	return nil
}

func main() {
	var flags flag.FlagSet

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		for _, f := range plugin.Files {
			err := generateGenericFriendlyEnumsForEventTypes(plugin, f)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
