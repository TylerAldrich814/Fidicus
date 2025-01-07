package proto

import (
	"strings"
	"sync"

	"github.com/bufbuild/protocompile/linker"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ProtoPortion string
const (
  ProtoMeta      ProtoPortion = "metadata"
  ProtoEnums     ProtoPortion = "enums"
  ProtoMsgDefs   ProtoPortion = "msgDefs"
  ProtoMsgParams ProtoPortion = "msgParams"
  ProtoSvcDefs   ProtoPortion = "svcDefs"
  ProtoSvcParams ProtoPortion = "svcParams"
  ProtoRelations ProtoPortion = "relations"
)

type ProtoCypherCompiler struct { 
  builders map[ProtoPortion]*strings.Builder
  metadata ProtoMetadata
  wg       sync.WaitGroup
}

func NewProtoCypherCompiler() *ProtoCypherCompiler {
  builders := map[ProtoPortion]*strings.Builder{
    ProtoMeta      : new(strings.Builder),
    ProtoEnums     : new(strings.Builder),
    ProtoMsgDefs   : new(strings.Builder),
    ProtoMsgParams : new(strings.Builder),
    ProtoSvcDefs   : new(strings.Builder),
    ProtoSvcParams : new(strings.Builder),
    ProtoRelations : new(strings.Builder),
  }

  return &ProtoCypherCompiler{
    builders : builders,
    wg       : sync.WaitGroup{},
  }
}

func(c *ProtoCypherCompiler) Run(
  dep *ProtoMetadata,
) error {
  file    := dep.File
  pkgName := dep.PkgName
  // imports := dep.Imports
  ver, pkg, err := extractVersion(pkgName)
  if err != nil {
    return err
  }
  enums    := []string{}
  messages := []string{}
  services := []string{}

  c.compileMetadata(file, ver, pkg, dep.Imports)
  c.compileEnums(file, ver, pkg, &enums)
  c.compileMessageDefinitions(file, ver, pkg, &messages)
  c.compileMessageParams(file, ver, pkg, )
  c.compileServiceDefinitions(file, ver, pkg, &services)
  c.compileServiceParams(file, ver, pkg)

  c.wg.Wait()
  for _, enum := range enums {
    appendQuery(c.builders[ProtoRelations],
      `(%s)-[:DEFINED_IN]->(%s),`,
      enum,
      pkgName,
    )
  }
  for _, msg := range messages {
    appendQuery(c.builders[ProtoRelations],
      `(%s)-[:DEFINED_IN]->(%s),`,
      msg,
      pkgName,
    )
  }
  for _, svc := range services {
    appendQuery(c.builders[ProtoRelations],
      `(%s)-[:DEFINED_IN]->(%s),`,
      svc,
      pkgName,
    )
  }
  for _, imp := range dep.Imports {
    appendQuery(c.builders[ProtoRelations],
    `(%s)-[:IMPORTS]->(%s);`,
      pkgName,
      imp,
    )
  }

  return nil
}


func(c *ProtoCypherCompiler) compileMetadata(
  file linker.File,
  ver  string,
  pkg  string,
  imps []string,
) {
  c.wg.Add(1)
  go func(){
    defer c.wg.Done()
    // pkgKey  := versionedKey(ver, pkg)
    pkgName := string(file.Package())
    syntax := file.Syntax().String()
    imports := ""
    if len(imps) != 0 {
      imports = strings.Join(imps, ", ")
    }

    appendQuery(c.builders[ProtoMeta],
`(%s:Package {
  name: "%s",
  package: "%s",
  syntax: "%s",
  imports: "%s"
}),`,
      pkgName,
      pkg,
      pkgName,
      syntax,
      imports,
    )
  }()
}

// compileEnums takes all parsed proto eneum data and compiles cypher queries
// for both enum nodes and relations.
func(c *ProtoCypherCompiler) compileEnums(
  file     linker.File,
  ver      string,
  pkg      string,
  enumKeys *[]string,
){
  c.wg.Add(1)
  go func(){
    defer c.wg.Done()
    var (
      enum       protoreflect.EnumDescriptor
      enumValue  protoreflect.EnumValueDescriptor
      opts       protoreflect.ProtoMessage
      descriptor protoreflect.MessageDescriptor
      usedIdxs   map[int32][]string
      enumName   string
      enumKey    string
      valueName   string
      valueNum   int32
      valueKey   string
      allowAlias bool
      deprecated bool
    )
    enums := file.Enums()

    for i := 0; i < enums.Len(); i++ {
      usedIdxs   = make(map[int32][]string)
      enum       = enums.Get(i)
      enumName   = string(enum.Name())
      enumKey    = versionedKey(ver, enumName)
      opts       = enum.Options()
      allowAlias = false
      deprecated = false
      *enumKeys  = append(*enumKeys, enumKey)

      // Enum options: allow_alias or deprecated
      if opts != nil {
        descriptor = opts.ProtoReflect().Descriptor()
        if field := descriptor.Fields().ByName("allow_alias"); field != nil {
          allowAlias = opts.ProtoReflect().Get(field).Bool()
        }
        if field := descriptor.Fields().ByName("deprecated"); field != nil {
          deprecated = opts.ProtoReflect().Get(field).Bool()
        }
      }
      appendQuery(c.builders[ProtoEnums],
`(%s:Enum {
  package: "%s",
  name: "%s",
  version: "%s",
  allowAlias: %t,
  deprecated: %t
}),`,
        enumKey,
        pkg,
        enumName,
        ver,
        allowAlias,
        deprecated,
      )

      for j := 0; j < enum.Values().Len(); j++ {
        enumValue   = enum.Values().Get(j)
        valueName   = string(enumValue.Name())
        valueNum    = int32(enumValue.Number())
        valueKey    = versionedKey(ver, enumName, valueName)

        if _, ok := usedIdxs[valueNum]; !ok {
          usedIdxs[valueNum] = []string{valueKey}
        } else {
          usedIdxs[valueNum] = append(usedIdxs[valueNum], valueKey)
        }

        appendQuery(c.builders[ProtoEnums],
`(%s:EnumValue {
  name: "%s",
  number: %d
}),`,
        valueKey,
        valueName,
        valueNum,
        )
      }
      // Compile Aliased Enum Value Relationships
      for _, v := range usedIdxs {
        if len(v) == 1 {
          continue
        }
        first := v[0]
        for _, next := range v[1:] {
          appendQuery(c.builders[ProtoEnums],
            `(%s)-[:ALIAS]->(%s),`,
            next,
            first,
          )
        }
      }
    }
  }()
}

func(c *ProtoCypherCompiler) compileMessageDefinitions(
  file linker.File,
  ver  string,
  pkg  string,
  msgs *[]string,
){
  c.wg.Add(1)
  go func(){
    defer c.wg.Done()
    var (
      descriptor  protoreflect.MessageDescriptor
      msg         protoreflect.MessageDescriptor
      opts        protoreflect.ProtoMessage
      msgName     string
      msgKey      string
      deprecated  bool
    )

    messages := file.Messages()
    for i := 0; i < messages.Len(); i++ {
      msg        = messages.Get(i)
      msgName    = string(msg.Name())
      msgKey     = versionedKey(ver, msgName)
      *msgs      = append(*msgs, msgKey)
      opts       = msg.Options()
      deprecated = false

      if opts != nil {
        descriptor = opts.ProtoReflect().Descriptor()
        if field := descriptor.Fields().ByName("deprecated"); field != nil {
          deprecated = opts.ProtoReflect().Get(field).Bool()
        }
      }
      appendQuery(c.builders[ProtoMsgDefs],
`(%s:Message {
  package: "%s",
  version: "%s",
  name: "%s",
  deprecated: %t
}),`,
        msgKey,
        pkg,
        ver,
        msgName,
        deprecated,
      )
    }
  }()
}

// compileMessageParams -- Compilse Cypher Queries that define Message Parameter
// Nodes and Relationships.
func(c *ProtoCypherCompiler) compileMessageParams(
  file linker.File,
  ver  string,
  pkg  string,
){
  c.wg.Add(1)
  go func(){
    defer c.wg.Done()
    var (
      msg           protoreflect.MessageDescriptor
      field         protoreflect.FieldDescriptor
      paramKey      string
      msgName       string
      msgKey        string
      fieldKind     string
      fieldPackage  string
      fieldName     string
      tKey          string
      tVal          string
      fieldNum      int32
      isRepeated    bool
      isOptional    bool
      isMap         bool
    )
    messages := file.Messages()

    for i := 0; i < messages.Len(); i++ {
      msg     = messages.Get(i)
      msgName = string(msg.Name())
      msgKey  = versionedKey(ver, msgName)
      for j := 0; j < msg.Fields().Len(); j++ {
        field         = msg.Fields().Get(j)
        fieldKind     = field.Kind().String()
        fieldName     = string(field.Name())
        fieldNum      = int32(field.Number())
        paramKey      = versionedKey(ver, msgName, fieldName)
        isRepeated    = false
        isOptional    = false

        switch field.Cardinality() {
        case protoreflect.Repeated:
          isRepeated = true
        case protoreflect.Optional:
          isOptional = true
        }

        isMap = field.IsMap()
        tKey  = ""
        tVal  = ""
        if isMap {
          tKey = field.MapKey().Kind().String()
          tVal = field.MapValue().Kind().String()
          fieldKind = "map"
        }

        appendQuery(c.builders[ProtoMsgParams],
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
          paramKey,
          pkg,
          msg.Name(),
          isRepeated,
          isOptional,
          fieldName,
          fieldKind,
          fieldNum,
          tKey,
          tVal,
        )
      }
      appendQuery(c.builders[ProtoMsgParams],
        `(%s)-[:HAS_PARAMTER]->(%s),`,
        msgKey,
        paramKey,
      )
      if field.Kind() == protoreflect.MessageKind && !isMap {
        fieldType := field.Message().Name()
        appendQuery(c.builders[ProtoMsgParams],
          `(%s)-[:USES_MSG_TYPE]->(%s),`,
          paramKey,
          versionedKey(ver, string(fieldType)),
        )

        // If field is of type message FROM a different package.
        // Create relationship between :Paramter and :Package
        fieldPackage = strings.Join(
          strings.Split(
            string(field.Message().FullName()), 
            ".",
          )[:2],
          ".",
        )
        if fieldPackage != pkg+"."+ver {
          appendQuery(c.builders[ProtoMsgParams],
            `(%s)-[:FROM_PACKAGE]->(%s),`,
            paramKey,
            fieldPackage+"."+ver,
          )
        }

      } else if field.Kind() == protoreflect.EnumKind {
        enumName := string(field.Enum().Name())
        appendQuery(c.builders[ProtoMsgParams],
        `(%s)-[:USES_ENUM_TYPE]->(%s),`,
          paramKey,
          versionedKey(ver, enumName),
        )
      }
    }
  }()
}

func(c *ProtoCypherCompiler) compileServiceDefinitions(
  file linker.File,
  ver  string,
  pkg  string,
  svcs *[]string,
){
  c.wg.Add(1)
  go func(){
    defer c.wg.Done()
    services := file.Services()
    var (
      svc      protoreflect.ServiceDescriptor
      svcName  string
      svcKey   string
    )
    for i := 0; i < services.Len(); i++ {
      svc     = services.Get(i)
      svcName = string(svc.Name())
      svcKey  = versionedKey(ver, svcName)
      *svcs   = append(*svcs, svcKey)

      appendQuery(c.builders[ProtoSvcDefs],
`(%s:Service {
  name: "%s",
  package: "%s",
  version: "%s"
}),`,
      svcKey,
      svcName,
      pkg,
      ver,
      )
    }
  }()
}

func(c *ProtoCypherCompiler) compileServiceParams(
  file linker.File,
  ver  string,
  pkg  string,
){
  c.wg.Add(1)
  go func(){
    defer c.wg.Done()
    services := file.Services()
    
    var (
      svc        protoreflect.ServiceDescriptor
      method     protoreflect.MethodDescriptor
      methodName string
      methodKey  string
      svcName    string
      svcKey     string
      inputKey   string
      outputKey  string
    )
    for i := 0; i < services.Len(); i++ {
      svc     = services.Get(i)
      svcName = string(svc.Name())
      svcKey  = versionedKey(ver, svcName)

      for j := 0; j < svc.Methods().Len(); j++ {
        method     = svc.Methods().Get(j)
        methodName = string(method.Name())
        methodKey  = versionedKey(ver, svcName, methodName)
        inputKey   = versionedKey(ver, string(method.Input().Name()))
        outputKey  = versionedKey(ver, string(method.Output().Name()))
      }
      appendQuery(c.builders[ProtoSvcParams],
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
      appendQuery(c.builders[ProtoSvcParams],
        `(%s)-[:INPUT]->(%s),`,
        inputKey,
        methodKey,
      )
      appendQuery(c.builders[ProtoSvcParams],
        `(%s)-[:OUTPUT]->(%s),`,
        outputKey,
        methodKey,
      )
      appendQuery(c.builders[ProtoSvcParams],
        `(%s)-[:RPC_METHOD]->(%s),`,
        methodKey,
        svcKey,
      )
    }
  }()
}

func wrapMergeCypher(cypher string) string {
  return "BEGIN\n"+
         "MERGE\n"+
         "%s\n"+
         "COMMIT\n"+
         cypher+
         "COMMIT\n"+
         "EXCEPTION\n"+
         "    WHEN ANY THEN ROLLBACK;\n"
}

// WriteString -- Pulls all of our compiled Cypher Queries and combines 
// then into a unified Cypher Query String.
func(c *ProtoCypherCompiler) WriteString() string {
  cy := []string{
    "BEGIN\nMERGE\n",
    c.builders[ProtoMeta].String(),
    c.builders[ProtoEnums].String(),
    c.builders[ProtoMsgDefs].String(),
    c.builders[ProtoMsgParams].String(),
    c.builders[ProtoSvcDefs].String(),
    c.builders[ProtoSvcParams].String(),
    c.builders[ProtoRelations].String(),
    "COMMIT\nEXCEPTION\n\tWHEN ANY THEN ROLLBACK;\n",
  }
  return strings.Join(cy, "")
  // cypher := fmt.Sprintf(
  //   "%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n",
  //   "BEGIN\nMERGE",
  //   c.builders[ProtoMeta],
  //   c.builders[ProtoEnums],
  //   c.builders[ProtoMsgDefs],
  //   c.builders[ProtoMsgParams],
  //   c.builders[ProtoSvcDefs],
  //   c.builders[ProtoSvcParams],
  //   c.builders[ProtoRelations],
  //   "COMMIT\nEXCEPTION\n  WHEN ANY THEN ROLLBACK;",
  // )
  // return cypher
}
