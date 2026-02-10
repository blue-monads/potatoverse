package cli

type PackageCmd struct {
	Build PackageBuildCmd `cmd:"" help:"Build the package."`
	Push  PackagePushCmd  `cmd:"" help:"Push the package."`
}

type PackagePushCmd struct {
	PotatoYamlFile string `name:"potato-yaml-file" help:"Path to potato manifest file." type:"path" default:"./potato.yaml"`
	OutputZipFile  string `name:"output-zip-file" help:"Output path for built package." type:"path"`
}

type PackageBuildCmd struct {
	PotatoYamlFile string `name:"potato-yaml-file" help:"Path to potato manifest file." type:"path" default:"./potato.yaml"`
	OutputZipFile  string `name:"output-zip-file" help:"Output path for built package." type:"path"`
}
