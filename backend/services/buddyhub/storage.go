package buddyhub

import (
	"os"
	"path"
)

func (h *BuddyHub) GetBuddyDir(buddyPubkey string) (string, error) {

	buddyDir := path.Join(h.baseBuddyDir, buddyPubkey)

	os.MkdirAll(buddyDir, 0755)

	return buddyDir, nil
}
