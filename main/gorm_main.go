package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 特别注意：结构体名称为：Product，创建的表的名称为：Products
type Product struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Number         string    `gorm:"unique" json:"number"`                       //商品编号（唯一）
	Category       string    `gorm:"type:varchar(256);not null" json:"category"` //商品类别
	Name           string    `gorm:"type:varchar(20);not null" json:"name"`      //商品名称
	MadeIn         string    `gorm:"type:varchar(128);not null" json:"made_in"`  //生产地
	ProductionTime time.Time `json:"production_time"`                            //生产时间
}

// GormResponse 响应体
type GormResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

var gormDB *gorm.DB
var gormResponse GormResponse

func init() {
	var err error
	databaseConStr := "root:root1234@tcp(127.0.0.1:3306)/db2019?charset=utf8&parseTime=true&loc=Local"
	gormDB, err = gorm.Open(mysql.Open(databaseConStr), &gorm.Config{}) //配置项中预设了连接池 ConnPool
	if err != nil {
		fmt.Println("数据库连接出现了问题：", err)
		return
	}

}

func main1() {
	r := gin.Default()
	//数据库的CRUD--->gin的 post、get、put、delete方法
	r.POST("gorm/insert", gormInsertData) //添加数据
	r.GET("gorm/get", gormGetData)        //查询数据（单条记录）
	// r.GET("gorm/mulget", gormGetMulData)    //查询数据（多条记录）
	// r.PUT("gorm/update", gormUpdateData)    //更新数据
	// r.DELETE("gorm/delete", gormDeleteData) //删除数据
	r.Run(":9090")
}
func gormGetData(c *gin.Context) {
	//=============捕获异常============
	defer func() {
		err := recover()
		if err != nil {
			gormResponse.Code = http.StatusBadRequest
			gormResponse.Message = "错误"
			gormResponse.Data = err
			c.JSON(http.StatusBadRequest, gormResponse)
		}
	}()
	//============
	number := c.Query("number")
	product := Product{}
	tx := gormDB.Where("number=?", number).First(&product)
	if tx.Error != nil {
		gormResponse.Code = http.StatusBadRequest
		gormResponse.Message = "查询错误"
		gormResponse.Data = tx.Error
		c.JSON(http.StatusOK, gormResponse)
		return
	}
	gormResponse.Code = http.StatusOK
	gormResponse.Message = "读取成功"
	gormResponse.Data = product
	c.JSON(http.StatusOK, gormResponse)
}

func gormInsertData(c *gin.Context) {
	//=============捕获异常============
	defer func() {
		err := recover()
		if err != nil {
			gormResponse.Code = http.StatusBadRequest
			gormResponse.Message = "错误"
			gormResponse.Data = err
			c.JSON(http.StatusBadRequest, gormResponse)
		}
	}()
	//============
	var p Product
	err := c.Bind(&p)
	if err != nil {
		gormResponse.Code = http.StatusBadRequest
		gormResponse.Message = "参数错误"
		gormResponse.Data = err
		c.JSON(http.StatusOK, gormResponse)
		return
	}
	fmt.Println(p)
	tx := gormDB.Create(&p)
	if tx.RowsAffected > 0 {
		gormResponse.Code = http.StatusOK
		gormResponse.Message = "写入成功"
		gormResponse.Data = "OK"
		c.JSON(http.StatusOK, gormResponse)
		return
	}
	fmt.Printf("insert failed, err:%v\n", err)
	gormResponse.Code = http.StatusBadRequest
	gormResponse.Message = "写入失败"
	gormResponse.Data = tx
	c.JSON(http.StatusOK, gormResponse)
	fmt.Println(tx) //打印结果
}
