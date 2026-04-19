package cli

type PackageCmd struct {
	Init  PackageInitCmd  `cmd:"" help:"Initialize a new project from a template."`
	Build PackageBuildCmd `cmd:"" help:"Build the package."`
	Push  PackagePushCmd  `cmd:"" help:"Push the package."`
}

type PackagePushCmd struct {
	PotatoYamlFile string `name:"potato-yaml-file" help:"Path to potato manifest file." type:"path" default:"./potato.yaml"`
	OutputZipFile  string `name:"output-zip-file" help:"Output path for built package." type:"path"`

	// InferValidate  bool   `name:"infer-validate" help:"Infer validate the package." type:"bool" default:"false"`
	// ValidateSQLFiles []string `name:"validate-sql-files" help:"Validate the SQL files." type:"list"`
	// ValidateLuaFiles []string `name:"validate-lua-files" help:"Validate the Lua files." type:"list"`
}

type PackageBuildCmd struct {
	PotatoYamlFile string `name:"potato-yaml-file" help:"Path to potato manifest file." type:"path" default:"./potato.yaml"`
	OutputZipFile  string `name:"output-zip-file" help:"Output path for built package." type:"path"`
}
