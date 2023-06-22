package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	connStr    = flag.String("conn_str", "", "Connect string of MySQL")
	listenAddr = flag.String("listen_addr", "", "Server listen address")
)

type Article struct {
	ID        uint      `gorm:"primary_key"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Name      string
	Content   []Paragraph
}

type Paragraph struct {
	ID        uint `gorm:"primary_key"`
	ArticleID uint
	Type      ParaType
	Content   string
	Metadata  string
}

type ParaType string

const (
	PRIMARY_HEADER   ParaType = "PRIMARY_HEADER"
	SECONDARY_HEADER ParaType = "SECONDARY_HEADER"
	TEXT             ParaType = "TEXT"
	CODE             ParaType = "CODE"
)

var db *gorm.DB

func main() {
	flag.Parse()

	var err error

	db, err = gorm.Open(mysql.Open(*connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	err = db.AutoMigrate(&Article{}, &Paragraph{})
	if err != nil {
		log.Fatalf("failed to migrate database schema: %v", err)
	}

	r := gin.Default()

	r.POST("/articles", createOrUpdateArticle)
	r.GET("/articles", getArticles)
	r.DELETE("/articles/:id", deleteArticle)

	err = r.Run(*listenAddr)
	if err != nil {
		panic(err.Error())
	}
}

func createOrUpdateArticle(c *gin.Context) {
	var article Article
	err := c.ShouldBindJSON(&article)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if article already exists
	result := db.Where("name = ?", article.Name).First(&article)
	if result.Error == nil {

		// Article exists, update it
		err = db.Model(&article).Association("Content").Replace(article.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	} else {

		// Article does not exist, create a new one
		err = db.Create(&article).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "created"})
	}
}

func getArticles(c *gin.Context) {
	var articles []Article

	// Preload paragraphs
	db.Preload("Content").Order("id DESC").Find(&articles)

	c.JSON(http.StatusOK, articles)
}

func deleteArticle(c *gin.Context) {
	id := c.Param("id")

	articleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article ID"})
		return
	}

	// Delete article
	result := db.Delete(&Article{ID: uint(articleID)})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
