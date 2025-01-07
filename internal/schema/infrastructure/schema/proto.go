package schema

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/TylerAldrich814/Fidicus/internal/shared/utils"
	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Proto struct { 
}

func NewProto() *Proto {
	return &Proto{  }
}

func loadProtoSources(root string) map[string]string {
  sources := make(map[string]string)
  _ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
    if err == nil && strings.HasSuffix(path, ".proto") {
      data, readErr := os.ReadFile(path)
      if readErr == nil {
        relPath, _ := filepath.Rel(root, path)
        sources[relPath] = string(data)
      }
    }
    return nil
  })

  return sources
}


// ParseSchemaFile - Taking in a filename -- Loads in the Proto schema file.
// Then parses all sematic information from the proto file and converts it 
// into a cypher query.
func (p *Proto) ParseSchemaFile(
  ctx    context.Context,
  files  []string,
  runner func(queryBuilder string) error,
) error {
  var pushLog = utils.NewLogHandlerFunc(
    "ParseSchemaFile",
    log.Fields{
      "files": "[ "+strings.Join(files, ", ")+" ]",
    },
  )

  // ProtoReflect Setup:
	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{
			Accessor: protocompile.SourceAccessorFromMap(loadProtoSources(".")), // Loads proto files from current directory
		},
	}

  // Compile ProtoFile with ProtoReflect:
  compiledFiles, err := compiler.Compile(ctx, files...)
  if err != nil {
    pushLog(
      utils.LogErro,
      "Failed to parse Proto: %v",
      err,
    )
    return ErrSchemaParseFailed
  }

  for _, file := range compiledFiles {
    builders, err := compileProtoFile(file)
    if err != nil {
      return err
    }
    for _, query := range builders {
      if err := runner(query.String()); err != nil {
        pushLog(
          utils.LogErro,
          "failed to run proto queries: %v",
          err,
        )
        return err
      }
    }
  }

  
  return nil
}

func compileProtoFile(
  file linker.File,
)( []*strings.Builder, error ){
  if file.Syntax().String() != "proto3" {
    return nil, fmt.Errorf("currently only support proto3: not %s", file.Syntax().String())
  }
  packageName := file.Package()
  ver, pkg, err := extractVersion(string(packageName))
  if err != nil {
    return nil, err
  }

  var (
    metadataBuiler  strings.Builder
    enumBuilder     strings.Builder
    msgDefBuilder   strings.Builder
    msgParamBuilder strings.Builder
    svcDefBuilder   strings.Builder
    svcMethBuilder  strings.Builder
    protoRelations  strings.Builder

    enums []string
    msgs  []string
    svcs  []string
  )
  title := fmt.Sprintf("------------------ %s ------------------ ", string(packageName))

  metadataBuiler.WriteString(fmt.Sprintf("  %s \n", title))

  builders := make([]*strings.Builder, 7)
  messages := file.Messages()
  services := file.Services()

  var wg sync.WaitGroup
  
  // 1: ->> Parse Package Metadata && create cypher queryBuilder for package node
  wg.Add(1)
  go func(){
    log.Printf("->Starting Metadata")
    defer wg.Done()
    _, _, err := processPackageMetadata(
      file,
      ver,
      pkg,
      &metadataBuiler,
    )
    if err != nil {
      log.Errorf("Failed to parse packge metadata for %s: %s",file.Package().Name(), err.Error())
    }
    builders[0] = &metadataBuiler
  }()

  // 2: ->> Process Enums Definitions:
  wg.Add(1)
    log.Printf("->Starting Enums")
  go func(){
    defer wg.Done()
    enums = processEnumHandles(
      file.Enums(), 
      pkg,ver, 
      &enumBuilder,
    )
    builders[1] = &enumBuilder
  }()

  // 3: ->> Process Messages Definitions:
  wg.Add(1)
  go func(){
    log.Printf("->Starting MsgDefs")
    defer wg.Done()
    msgs = processMessages(
      messages, 
      pkg, 
      ver, 
      &msgDefBuilder,
    )
    builders[2] = &msgDefBuilder
  }()

  // 4:->> Process Message Paramters:
  wg.Add(1)
  go func(){
    log.Printf("->Starting MsgParams")
    defer wg.Done()
    processMessageParameters(
      messages, 
      pkg, 
      ver, 
      &msgParamBuilder,
    )
    builders[3] = &msgParamBuilder
  }()

  // 5: ->> Process svc Definitions:
  wg.Add(1)
  go func(){
    log.Printf("->Starting SvcDefs")
    defer wg.Done()
    svcs = processServiceDefinition(
      services, 
      pkg, 
      ver, 
      &svcDefBuilder,
    )
    builders[4] = &svcDefBuilder
  }()

  // 6: ->> Process rfc Definitions:
  wg.Add(1)
  go func(){
    log.Printf("->Starting RFCDefs")
    defer wg.Done()
    processServiceMethods(
      services, 
      pkg, 
      ver, 
      &svcMethBuilder,
    )
    builders[5] = &svcMethBuilder
  }()

  wg.Wait()

  log.Printf("->Starting Relations:")
  for _, e := range enums {
    appendQuery(&protoRelations,
      `(%s)-[:DEFINED_IN]->(%s),`,
      e,
      packageName,
    )
  }
  for _, m := range msgs {
    appendQuery(&protoRelations,
      `(%s)-[:DEFINED_IN]->(%s),`,
      m,
      packageName,
    )
  }

  for _, s := range svcs {
    appendQuery(&protoRelations,
      `(%s)-[:DEFINED_IN]->(%s),`,
      s,
      packageName,
    )
  }
  protoRelations.WriteString(fmt.Sprintf(" %s \n", strings.Repeat("-", len(title))))

  builders[6] = &protoRelations

  return builders, nil
}

func processPackageMetadata(
  file    linker.File,
  ver     string,
  pkg     string,
  queryBuilder *strings.Builder,
)( string, []string, error){
  packageName   := string(file.Package())
  syntax        := file.Syntax().String()

  imports := []string{}
  for i := 0; i < file.Imports().Len(); i++ {
    imp := file.Imports().Get(i)
    impPackage := imp.Package()
    imports = append(imports, string(impPackage))
  }

  appendQuery(queryBuilder,
`(%s:Package { 
  name: "%s", 
  package: "%s", 
  version: "%s",
  syntax: "%s", 
  imports: [%s] 
}),`,
    packageName,
    packageName,
    pkg,
    ver,
    syntax,
    strings.Join(imports, ", "),
  )

  return packageName, imports, nil
}

func processEnumHandles(
  enums   protoreflect.EnumDescriptors,
  pkg     string,
  ver     string,
  queryBuilder *strings.Builder,
) []string {
  var (
    enum        protoreflect.EnumDescriptor
    enumValue   protoreflect.EnumValueDescriptor
    opts        protoreflect.ProtoMessage
    descriptor  protoreflect.MessageDescriptor
    val         protoreflect.Value
    usedIndexes map[int32][]string
    enumName    string
    enumKey     string
    valueName   string
    valueKey    string
    valueNumber int32
    allowAlias  bool
    deprecated  bool
  )
  enumKeys   := []string{}
  for i := 0; i < enums.Len(); i++ {
    usedIndexes = make(map[int32][]string)
    enum        = enums.Get(i)
    enumName    = string(enum.Name())
    enumKey     = versionedKey(ver, enumName)
    opts        = enum.Options()

    allowAlias = false
    deprecated = false

    enumKeys = append(enumKeys, enumKey)

    if opts != nil {
      descriptor = opts.ProtoReflect().Descriptor()

      // Parse : option allow_alias = <BOOL>
      if field := descriptor.Fields().ByName("allow_alias"); field != nil {
        val = opts.ProtoReflect().Get(field)
        allowAlias = val.Bool()
      }
      // Parse : option deprecated = <BOOL>
      if field := descriptor.Fields().ByName("deprecated"); field != nil {
        val = opts.ProtoReflect().Get(field)
        deprecated = val.Bool()
      }
    }
    appendQuery(queryBuilder,
`(%s:Enum { 
  package: "%s", 
  name: "%s", 
  ver: "%s", 
  allowAlias: %t, 
  deprecated: %t 
}),`,
      enumKey,
      pkg,
      enumKey,
      ver,
      allowAlias,
      deprecated,
    )

    // ->> Compile Enum Fields:
    for i := 0; i < enum.Values().Len(); i++ {
      enumValue       = enum.Values().Get(i)
      valueName   = string(enumValue.Name())
      valueNumber = int32(enumValue.Number())
      valueKey    = versionedKey(ver, enumName, valueName)

      if _, ok := usedIndexes[valueNumber]; !ok {
        usedIndexes[valueNumber] = []string{valueKey}
      } else {
        usedIndexes[valueNumber] = append(usedIndexes[valueNumber], valueKey)
      }

      appendQuery(queryBuilder,
`(%s:EnumValue { 
  name: "%s", 
  number: %d 
}),`,
        valueKey,
        valueName,
        valueNumber,
      )
    }
    // Compile Aliased Enum Relationships:
    for _, v := range usedIndexes {
      if len(v) > 1 {
        first := v[0]
        for _, next := range v[1:] {
          appendQuery(queryBuilder,
            fmt.Sprintf(`(%s)-[:ALIAS]->(%s),`,
            next,
            first,
          ))
        }
      }
    }
  }
  return enumKeys
}


func processMessages(
  messages protoreflect.MessageDescriptors,
  pkg      string,
  ver      string,
  queryBuilder *strings.Builder,
) []string {
  msgs := []string{}
  var (
    descriptor  protoreflect.MessageDescriptor
    msg         protoreflect.MessageDescriptor
    opts        protoreflect.ProtoMessage
    val         protoreflect.Value
    msgName     string
    msgKey      string
    deprecated  bool
  )
  for i := 0; i < messages.Len(); i++ {
    msg        = messages.Get(i)
    msgName    = string(msg.Name()) 
    msgKey     = versionedKey(ver, msgName)
    msgs       = append(msgs, msgKey)
    opts       = msg.Options()
    deprecated = false

    if opts != nil {
      descriptor = opts.ProtoReflect().Descriptor()

      if field := descriptor.Fields().ByName("deprecated"); field != nil {
        val = opts.ProtoReflect().Get(field)
        deprecated = val.Bool()
      }
    }

    appendQuery(queryBuilder,
`(%s:Message { 
  package: "%s", 
  version: "%s", 
  name: "%s", 
  deprecated: %t}
),`,
      msgKey,
      pkg,
      ver,
      msgName,
      deprecated,
    )
  }
  return msgs
}

func processMessageParameters(
  msgs    protoreflect.MessageDescriptors,
  pkg     string,
  ver     string,
  queryBuilder *strings.Builder,
) {
  var (
    msg         protoreflect.MessageDescriptor
    field       protoreflect.FieldDescriptor
    cardinality protoreflect.Cardinality
    msgName     string
    msgKey      string
    fieldKind   string
    fieldName   string
    fieldNum    int32
  )
  for i := 0; i < msgs.Len(); i++ {
    msg     = msgs.Get(i)
    msgName = string(msg.Name())
    msgKey  = versionedKey(ver, msgName)
    for i := 0; i < msg.Fields().Len(); i++ {
      field       = msg.Fields().Get(i)
      fieldKind   = field.Kind().String()
      fieldNum    = int32(field.Number())
      fieldName   = string(field.Name())
      cardinality = field.Cardinality()

      var (
        isRepeated bool = false
        isOptional bool = false
      )
      switch cardinality {
      case protoreflect.Repeated:
        isRepeated = true
      case protoreflect.Optional:
        isOptional = true
      }

      isMap := field.IsMap()
      tKey := ""
      tVal := ""
      if isMap {
        tKey = field.MapKey().Kind().String()
        tVal = field.MapValue().Kind().String()

        // with proto3. When a Map<TYPE,TYPE> is used. Proto will actually create 
        // a hidden Message Type of <NAME>Entry with both key and val as parameters.
        // here, we'll instead just change the type from 'message' to 'map'
        fieldKind = "map"
      }

      mKey := versionedKey(ver, msgName, fieldName)
      appendQuery(queryBuilder,
`(%s:Parameter { 
  package: "%s", 
  message: "%s", 
  repeated: %t, 
  optional: %t, 
  field: "%s", 
  type: "%s", 
  number: %d, 
  tKey: "%s", 
  tVal: "%s"
}),`,
        mKey,
        pkg,
        string(msg.Name()),
        isRepeated,
        isOptional,
        fieldName,
        fieldKind,
        fieldNum,
        tKey, 
        tVal,
      )
      appendQuery(queryBuilder,
        `(%s)-[:HAS_PARAMETER]->(%s),`,
        msgKey,
        mKey,
      )

      if field.Kind() == protoreflect.MessageKind && !isMap {
        fieldType := field.Message().Name()
        appendQuery(queryBuilder,
        `(%s)-[:USES_MSG_TYPE]->(%s),`,
          mKey,
          string(fieldType) + "_" + ver,
        )
      } else if field.Kind() == protoreflect.EnumKind {
        enumName := string(field.Enum().Name()) + "_" + ver
        appendQuery(queryBuilder,
          `(%s)-[:USES_ENUM_TYPE]->(%s),`,
          mKey,
          enumName,
        )
      }
    }
  }
}


func processServiceDefinition(
  services protoreflect.ServiceDescriptors,
  pkg     string,
  ver     string,
  queryBuilder *strings.Builder,
) []string {
  svcs := []string{}
  for i := 0; i < services.Len(); i++{
    svc := services.Get(i)
    svcName := string(svc.Name()) + "_" + ver
    svcs = append(svcs, svcName)
    appendQuery(queryBuilder,
`(%s:Service { 
  name: "%s", 
  package: "%s", 
  version: "%s"
}),`,
      svcName,
      svcName,
      pkg,
      ver,
    )
  }
  return svcs
}

func processServiceMethods(
  services protoreflect.ServiceDescriptors,
  pkg     string,
  ver     string,
  queryBuilder *strings.Builder,
) {
  var (
   svc        protoreflect.ServiceDescriptor
   method     protoreflect.MethodDescriptor
   methodName string
   methodKey  string
   svcName    string
   svcKey     string
   inputMsg   string
   outputMsg  string
  )
  for i := 0; i < services.Len(); i++ {
    svc     = services.Get(i)
    svcName = string(svc.Name())
    svcKey  = versionedKey(ver, svcName)
    for j := 0; j < svc.Methods().Len(); j++ {
      method     = svc.Methods().Get(j)
      methodName = string(method.Name())
      inputMsg   = versionedKey(ver, string(method.Input().Name()))
      outputMsg  = versionedKey(ver, string(method.Output().Name()))
      methodKey  = versionedKey(ver, svcName, methodName)

      // Create rpc method, define Input & Output message relationships.
      appendQuery(queryBuilder,
`(%s:Method { 
  name: "%s", 
  package: "%s",
  version: "%s"
}),`,
        methodKey,
        methodName,
        pkg,
        ver,
      )
      appendQuery(queryBuilder,
      `(%s)-[:INPUT]->(%s),`,
        inputMsg,
        methodKey,
      )
      appendQuery(queryBuilder,
      `(%s)-[:OUTPUT]->(%s),`,
        outputMsg,
        methodKey,
      )

      appendQuery(queryBuilder,
         `(%s)-[:RPC_METHOD]->(%s),`,
        methodKey,
        svcKey,
      )
    }
  }
}

func appendQuery(
  queryBuilder *strings.Builder,
  f string,
  args ...any,
) {
  queryBuilder.WriteString(fmt.Sprintf(f+"\n", args...))
}

func extractVersion(
  packageName string,
)( string, string, error ){
  parts := strings.Split(packageName, ".")
  
  if strings.ToUpper(parts[len(parts)-1][0:1]) != "V" {
    return "", "", fmt.Errorf(
      "Not a supported versioning system. Version must be at the end of package name; Seperated by a '.' and appended with a 'v'|'V'",
    )
  }

  return parts[len(parts)-1], strings.Join(parts[:len(parts)-1], "."), nil
}

func versionedKey(version string, args ...string) string {
  return fmt.Sprintf("%s_%s", strings.Join(args, "_"), version)
}
