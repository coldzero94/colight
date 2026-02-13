package controller

import (
	"net/http"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type CrawlingController struct {
	crawlingService *service.CrawlingService
}

func NewCrawlingController(crawlingService *service.CrawlingService) *CrawlingController {
	return &CrawlingController{crawlingService: crawlingService}
}

// ParseJobPosting handles POST /v1/crawl
func (c *CrawlingController) ParseJobPosting(ctx *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "URL이 필요합니다.", "code": "VALID_001"},
		})
		return
	}

	result, err := c.crawlingService.CrawlJobPosting(ctx.Request.Context(), req.URL)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "크롤링 실패: " + err.Error(), "code": "SYS_001"},
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
