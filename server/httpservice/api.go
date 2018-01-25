package main

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"net/http"
)

func getRouter() *gin.Engine{

	router := gin.Default()

	router.PUT("/hashes/add/:hash", func(c *gin.Context){
		hash, err := strconv.ParseUint(c.Param("hash"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = RequestAddHash([]uint64{hash})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true} )
	})

	router.GET("/hashes/search/:hash/:distance", func(c *gin.Context){
		hash, err := strconv.ParseUint(c.Param("hash"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		distance, err := strconv.ParseUint(c.Param("distance"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		hashes, elapsed, err := RequestSearch(hash, uint32(distance))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"hashes": hashes,
			"elapsed_time": elapsed.Seconds(),
		})



	})

	return router


}