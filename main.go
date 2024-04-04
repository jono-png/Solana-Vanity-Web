package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gagliardetto/solana-go"
    "net/http"
    "regexp"
    "strings"
    "sync/atomic"
    "time"
)

var (
    generatedCount uint64
)

func validatePrefix(prefix string) bool {
    regex := regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]+$`)
    return regex.MatchString(prefix)
}

r.GET("/generation-count", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"count": generatedCount})
})

func generateWallet(prefix string) (string, string, time.Duration) {
    startTime := time.Now()
    for {
        newWallet := solana.NewWallet()
        publicKey := newWallet.PublicKey().String()
        if strings.HasPrefix(publicKey, prefix) {
            duration := time.Since(startTime)
            privateKeyStr := newWallet.PrivateKey.String()

            return publicKey, privateKeyStr, duration
        }
        atomic.AddUint64(&generatedCount, 1)
    }
}

func main() {
    r := gin.Default()

    r.Use(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusOK)
            return
        }
        c.Next()
    })

    r.GET("/start-generation", func(c *gin.Context) {
        prefix := c.Query("prefix")
        if !validatePrefix(prefix) {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prefix. Use 1-9, A-H, J-N, P-Z, a-k, m-z excluding 'I', 'O', 'l'."})
            return
        }
        walletAddress, privateKey, duration := generateWallet(prefix)
        c.JSON(http.StatusOK, gin.H{
            "wallet":     walletAddress,
            "privateKey": privateKey,
            "duration":   duration.String(),
            "attempts":   generatedCount,
        })
        atomic.StoreUint64(&generatedCount, 0)
    })

    r.Run(":8080")
}
