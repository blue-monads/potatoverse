package buddycdc

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazymodel"
)

const TemplateTable = `
CREATE TABLE IF NOT EXISTS {{.TableName}} (
    id INTEGER PRIMARY KEY,
	record_id INTEGER NOT NULL,
	operation INTEGER NOT NULL, -- 0: insert, 1: update, 2: delete
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

func (b *BuddyCDC) applyMeta(meta *lazymodel.BuddyCDCMeta) {

}
