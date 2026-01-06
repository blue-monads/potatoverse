package buddyhub

import (
	"os"
	"path"
)

func (h *BuddyHub) GetBuddyRoot(buddyPubkey string) (*os.Root, error) {

	buddyDir := path.Join(h.baseBuddyDir, buddyPubkey)

	os.MkdirAll(buddyDir, 0755)

	root, err := os.OpenRoot(buddyDir)
	if err != nil {
		return nil, err
	}

	return root, nil
}
