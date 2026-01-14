package buddyhub

import (
	"github.com/blue-monads/potatoverse/backend/services/buddyhub/funnel"
	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/blue-monads/potatoverse/backend/xtypes/buddy"
)

func (h *BuddyHub) startRloop() {
	for _, rendezvousUrl := range h.configuration.rendezvousUrls {
		go h.rLoopHandle(&rendezvousUrl)
	}
}

func (h *BuddyHub) rLoopHandle(rendezvousUrl *buddy.RendezvousUrl) {

	if rendezvousUrl.Provider != "funnel" {
		return
	}

	client := funnel.NewFunnelClient(funnel.FunnelClientOptions{
		LocalHttpPort:   h.port,
		RemoteFunnelUrl: rendezvousUrl.URL,
		ServerId:        h.pubkey,
	})

	token, err := xutils.GenerateNostrAuthToken(h.privkey, rendezvousUrl.URL, "GET")
	if err != nil {
		h.logger.Error("Failed to generate nostr auth token", "err", err)
		return
	}

	defer client.Stop()

	for {
		err = client.Start(token)
		if err != nil {
			h.logger.Error("Failed to start funnel client", "err", err)
			return
		}
	}

}
