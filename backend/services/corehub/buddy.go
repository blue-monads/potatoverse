package corehub

import (
	"os"
	"path"
)

func (c *CoreHub) GetBuddyRoot(nodeId string) (*os.Root, error) {

	buddyDir := path.Join(c.buddyDir, nodeId)

	os.MkdirAll(buddyDir, 0755)

	root, err := os.OpenRoot(buddyDir)
	if err != nil {
		return nil, err
	}

	return root, nil
}
