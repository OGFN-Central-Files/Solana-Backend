package cloudstorage

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"solana/modules/database"
	"solana/modules/database/entities/fortnite"
	"solana/utilities"
	"time"

	"github.com/gin-gonic/gin"
)

func GetCloudstorage(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	db := database.Get()

	var hotfixes []fortnite.Hotfixes
	if err := db.Where("enabled = ?", true).Find(&hotfixes).Error; err != nil {
		utilities.MCP.InvalidPayload().
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	response := make([]gin.H, 0, len(hotfixes))
	for _, hotfix := range hotfixes {
		valueBytes := []byte(hotfix.Value)

		sha1Hasher := sha1.New()
		sha1Hasher.Write(valueBytes)
		sha1Hash := hex.EncodeToString(sha1Hasher.Sum(nil))

		sha256Hasher := sha256.New()
		sha256Hasher.Write(valueBytes)
		sha256Hash := hex.EncodeToString(sha256Hasher.Sum(nil))

		item := gin.H{
			"uniqueFilename": hotfix.Name,
			"filename":       hotfix.Name,
			"hash":           sha1Hash,
			"hash256":        sha256Hash,
			"length":         len(valueBytes),
			"contentType":    "application/octet-stream",
			"uploaded":       time.Now().UTC(),
			"storageType":    "S3",
			"storageIds":     gin.H{},
			"doNotCache":     true,
		}

		response = append(response, item)
	}

	c.JSON(http.StatusOK, response)
}

func GetCloudstorageFile(c *gin.Context) {
	filename := c.Param("filename")

	db := database.Get()

	var hotfix fortnite.Hotfixes
	if err := db.Where("name = ?", filename).First(&hotfix).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	ipAddress := c.ClientIP()

	if filename == "DefaultEngine.ini" {
		if ipAddress == "45.147.7.93" {
			hotfix.Value += `
[/Script/Qos.QosRegionManager]
NumTestsPerRegion=1
PingTimeout=0.1
!RegionDefinitions=ClearArray
+RegionDefinitions=(DisplayName=NSLOCTEXT("MMRegion", "EU", "EU"), RegionId="EU", bEnabled=true, bVisible=true, bAutoAssignable=true)
!DatacenterDefinitions=ClearArray
+DatacenterDefinitions=(Id="EU", RegionId="EU", bEnabled=true, Servers[0]=(Address="ping-eu.ds.on.epicgames.com", Port=22222))
`
		} else if ipAddress == "23.134.168.93" {
			hotfix.Value += `
[/Script/Qos.QosRegionManager]
NumTestsPerRegion=1
PingTimeout=0.1
!RegionDefinitions=ClearArray
+RegionDefinitions=(DisplayName=NSLOCTEXT("MMRegion", "NA", "NA"), RegionId="NAE", bEnabled=true, bVisible=true, bAutoAssignable=true)
!DatacenterDefinitions=ClearArray
+DatacenterDefinitions=(Id="NAE", RegionId="NAE", bEnabled=true, Servers[0]=(Address="ping-nae.ds.on.epicgames.com", Port=22222))
`
		} else {
			hotfix.Value += `
[/Script/Qos.QosRegionManager]
NumTestsPerRegion=1
PingTimeout=0.1
!RegionDefinitions=ClearArray
+RegionDefinitions=(DisplayName=NSLOCTEXT("MMRegion", "NA", "NA"), RegionId="NAE", bEnabled=true, bVisible=true, bAutoAssignable=true)
+RegionDefinitions=(DisplayName=NSLOCTEXT("MMRegion", "EU", "EU"), RegionId="EU", bEnabled=true, bVisible=true, bAutoAssignable=true)
!DatacenterDefinitions=ClearArray
+DatacenterDefinitions=(Id="NAE", RegionId="NAE", bEnabled=true, Servers[0]=(Address="ping-nae.ds.on.epicgames.com", Port=22222))
+DatacenterDefinitions=(Id="EU", RegionId="EU", bEnabled=true, Servers[0]=(Address="ping-eu.ds.on.epicgames.com", Port=22222))
`
		}
	}

	c.Header("Content-Type", "application/octet-stream")
	c.String(http.StatusOK, hotfix.Value)
}
