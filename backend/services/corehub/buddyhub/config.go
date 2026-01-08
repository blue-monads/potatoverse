package buddyhub

import "github.com/blue-monads/turnix/backend/xtypes"

func (h *BuddyHub) configure(config *xtypes.AppOptions) error {
	buddyOptions := config.BuddyOptions

	if buddyOptions == nil {
		return nil
	}

	if len(h.staticBuddies) > 0 {
		for _, buddyInfo := range buddyOptions.StaticBuddies {
			h.staticBuddies[buddyInfo.Pubkey] = buddyInfo
		}
	}

	h.configuration = Configuration{
		allowAllBuddies:         buddyOptions.AllowAllBuddies,
		allbuddyAllowStorage:    buddyOptions.AllBuddyAllowStorage,
		allbuddyMaxStorage:      buddyOptions.AllBuddyMaxStorage,
		buddyWebFunnelMode:      buddyOptions.BuddyWebFunnelMode,
		allbuddyMaxTrafficLimit: buddyOptions.AllBuddyMaxTrafficLimit,
		rendezvousUrls:          buddyOptions.RendezvousUrls,
	}

	return nil

}
