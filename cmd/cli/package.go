package cli

type PackageCmd struct {
	Build PackageBuildCmd `cmd:"" help:"Build the package."`
	Push  PackagePushCmd  `cmd:"" help:"Push the package."`
}

type PackagePushCmd struct {
	PotatoTomlFile string `name:"potato-toml-file" help:"Path to package directory." type:"path" default:"./potato.toml"`
	OutputZipFile  string `name:"output-zip-file" help:"Output path for built package." type:"path"`
}

type PackageBuildCmd struct {
	PotatoTomlFile string `name:"potato-toml-file" help:"Path to package directory." type:"path" default:"./potato.toml"`
	OutputZipFile  string `name:"output-zip-file" help:"Output path for built package." type:"path"`
}
