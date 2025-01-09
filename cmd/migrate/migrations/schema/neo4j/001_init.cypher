// -> Proto Files
CREATE INDEX FOR (n:Package) on (n.name, n.version);
CREATE INDEX FOR (n:Enum) ON (n.name, n.allowAlias, n.deprecated);
CREATE INDEX FOR (n:Message) ON (n.name);
CREATE INDEX FOR (n:Parameter) ON (n.message, n.field);
CREATE INDEX FOR (n:Service) ON (n.name);
CREATE INDEX FOR (n:Method) ON (n.name);
