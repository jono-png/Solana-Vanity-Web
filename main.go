package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gin-gonic/gin"
)

var (
	generatedCount    uint64
	shouldStopThreads atomic.Bool
	searchTerm        string
	numThreads        = 16
	logs              []string
	logsMutex         sync.Mutex
)

func addLog(log string) {
	logsMutex.Lock()
	defer logsMutex.Unlock()
	logs = append(logs, log)
}

func generateWallet(searchTerm string, startTime time.Time) {
	for !shouldStopThreads.Load() {
		newWallet := solana.NewWallet()
		pubKey := newWallet.PublicKey().String()
		if strings.HasPrefix(pubKey, searchTerm) {
			log := fmt.Sprintf("Success! Wallet found: %s\n", pubKey)
			addLog(log)
			shouldStopThreads.Store(true)
			break
		}
		atomic.AddUint64(&generatedCount, 1)
	}
}

func isValidPrefix(prefix string) bool {
	return len(prefix) <= 3 && strings.Trim(prefix, "A-HJ-NP-Za-km-z1-9") == ""
}

func main() {
	router := gin.Default()

	router.POST("/start-generation", func(c *gin.Context) {
		var json struct {
			Prefix string `json:"prefix"`
		}
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		searchTerm = json.Prefix
		if !isValidPrefix(searchTerm) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prefix"})
			return
		}

		shouldStopThreads.Store(false)
		generatedCount = 0
		logs = nil // reset logs
		startTime := time.Now()

		for i := 0; i < numThreads; i++ {
			go generateWallet(searchTerm, startTime)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Generation started"})
	})

	router.GET("/logs", func(c *gin.Context) {
		logsMutex.Lock()
		defer logsMutex.Unlock()
		c.JSON(http.StatusOK, gin.H{"logs": logs})
	})

	router.Run(":8080")
}
