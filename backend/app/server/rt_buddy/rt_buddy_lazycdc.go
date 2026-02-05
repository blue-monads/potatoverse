package rtbuddy

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	MaxLazyCDCBatchSize = 100
)

func (b *BuddyRouteServer) handleBuddyLazyCDCInit(ctx *gin.Context) {
	b.handleBuddyLazyCDCSyncMeta(ctx)
}

/*

todo => encrypt tablename and records



*/

func (b *BuddyRouteServer) handleBuddyLazyCDCSyncMeta(ctx *gin.Context) {
	_, err := verifyNostrAuthCtx(ctx, BuddyAuthExpiry)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	tables, err := b.selfcdc.GetAllCdcMeta()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"tables": tables})
}

type SyncRequest struct {
	TableId      int64 `json:"table_id"`
	LastSyncedId int64 `json:"last_synced_id"`
	Limit        int64 `json:"limit"`
}

func (b *BuddyRouteServer) handleBuddyLazyCDCSyncRecordSerial(ctx *gin.Context) {
	_, err := verifyNostrAuthCtx(ctx, BuddyAuthExpiry)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req SyncRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := req.Limit
	if limit <= 0 || limit > MaxLazyCDCBatchSize {
		limit = MaxLazyCDCBatchSize
	}

	records, err := b.selfcdc.GetTableRecordsSerial(req.TableId, req.LastSyncedId, int64(limit))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"records": records,
	})

}

func (b *BuddyRouteServer) handleBuddyLazyCDCSyncRecordCDC(ctx *gin.Context) {

}
