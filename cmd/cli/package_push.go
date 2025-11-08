package cli

import (
	"github.com/alecthomas/kong"
)

func (c *PackagePushCmd) Run(_ *kong.Context) error {

	zip, err := PackageFiles(c.PotatoTomlFile, c.OutputZipFile)
	if err != nil {
		return err
	}

	return PushPackage(c.PotatoTomlFile, zip)
}
