package rtbuddy

import (
	"net/http"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer"
	"github.com/gin-gonic/gin"
)

const (
	MaxLazyCDCBatchSize = 100
)

func (b *BuddyRouteServer) handleBuddyLazySyncMeta(ctx *gin.Context) {
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

func (b *BuddyRouteServer) handleBuddyLazySyncData(ctx *gin.Context) {
	_, err := verifyNostrAuthCtx(ctx, BuddyAuthExpiry)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req lazysyncer.SyncRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := req.Limit
	if limit <= 0 || limit > MaxLazyCDCBatchSize {
		limit = MaxLazyCDCBatchSize
	}

	bdata, err := b.selfcdc.GetDataCDC(req.TableId, req.LastSyncedId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, bdata)

}
