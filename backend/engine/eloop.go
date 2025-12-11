package engine

import "time"

func (e *Engine) startEloop() {

	readAllPendingPackageIds := func() []int64 {
		spaceIds := make([]int64, 0)

	main:
		for {

			select {
			case sid := <-e.reloadPackageIds:
				spaceIds = append(spaceIds, sid)
			default:
				break main
			}
		}
		return spaceIds
	}

	pendingFullReload := func() bool {
		for {
			select {
			case <-e.fullReload:
				return true
			default:
				return false
			}
		}
	}

	sTimer := time.NewTicker(time.Second * 2)
	defer sTimer.Stop()

	for range sTimer.C {
		if pendingFullReload() {
			e.loadRoutingIndex()
			readAllPendingPackageIds()
			continue
		}

		packageIds := readAllPendingPackageIds()
		if len(packageIds) > 0 {
			e.loadRoutingIndexForPackages(packageIds...)
		}

	}

}
