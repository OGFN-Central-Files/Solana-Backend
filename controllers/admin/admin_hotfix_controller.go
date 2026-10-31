package admin

import (
    "io/ioutil"
    "solana/modules/database"
    "solana/modules/database/entities/fortnite"
    "solana/utilities"
    "net/http"
    "github.com/gin-gonic/gin"
)

func GetHotfixes(c *gin.Context) {
    c.Header("Content-Type", "application/json")
    db := database.Get()
    var hotfixes []fortnite.Hotfixes
    if err := db.Where("enabled = ?", true).Find(&hotfixes).Error; err != nil {
        utilities.Internal.ServerError().Apply(c.Writer)
        return
    }
    var names []string
    for _, hotfix := range hotfixes {
        names = append(names, hotfix.Name)
    }
    c.JSON(http.StatusOK, names)
}

func GetHotfix(c *gin.Context) {
    file := c.Param("file")
    db := database.Get()
    var hotfix fortnite.Hotfixes
    if err := db.Where("name = ?", file).First(&hotfix).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Hotfix not found"})
        return
    }
    c.Data(http.StatusOK, "application/json", []byte(hotfix.Value))
}

func UpdateHotfix(c *gin.Context) {
    file := c.Param("file")
    db := database.Get()
    var hotfix fortnite.Hotfixes
    if err := db.Where("name = ?", file).First(&hotfix).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Hotfix not found"})
        return
    }
    body, err := ioutil.ReadAll(c.Request.Body)
    if err != nil {
        utilities.Internal.ServerError().Apply(c.Writer)
        return
    }
    hotfix.Value = string(body)
    if err := db.Save(&hotfix).Error; err != nil {
        utilities.Internal.ServerError().Apply(c.Writer)
        return
    }
    c.Data(http.StatusOK, "application/json", []byte(hotfix.Value))
}