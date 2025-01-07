package proto

import (
	"fmt"
	"strings"
)

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
