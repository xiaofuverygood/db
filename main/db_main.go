package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
 原生数据库操作
*/

var sqlDb *sql.DB           //数据库连接
var sqlResponse SqlResponse //响应结构体

/*
项目初始化
*/
func init() {
	/*
		连接数据库
		parseTime:时间格式转换(查询结果为时间时，是否自动解析为时间);
		loc=Local：MySQL的时区设置
	*/
	databaseConStr := "root:root1234@tcp(127.0.0.1:3306)/db2019?charset=utf8&parseTime=true&loc=Local"
	var err error
	sqlDb, err = sql.Open("mysql", databaseConStr)
	if err != nil {
		fmt.Println("连接数据库失败:", err)
		return
	}

	// 测试与数据库建立的连接 Ping()校验连接是否正确
	err = sqlDb.Ping()
	if err != nil {
		fmt.Println("连接数据库失败:", err)
		return
	}
}

// SqlUser 请求结构体
type SqlUser struct {
	Id      int    `json:"id"`
	Account string `json:"account"`
	Age     int    `json:"age"`
	Address string `json:"address"`
}

// SqlResponse 响应结构体
type SqlResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func main() {
	r := gin.Default()

	r.Static("/static", "./static/images")

	fs := http.Dir("static")
	fileHandler := http.FileServer(fs)
	http.Handle("/static/", http.StripPrefix("/static/", fileHandler))

	// 设置 CORS 头
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	//数据库的CRUD--->gin的 post、get、put、delete方法
	r.POST("sql/insert", insertData)   //添加数据
	r.GET("sql/get", getData)          //查询数据（单条记录）
	r.GET("sql/mulget", getMulData)    //查询数据（多条记录）
	r.PUT("sql/update", updateData)    //更新数据
	r.DELETE("sql/delete", deleteData) //删除数据
	r.Run(":9090")
}

/*
删除
*/
func deleteData(c *gin.Context) {
	account := c.Query("account")
	var count int

	//查询
	sqlStr := "SELECT count(*) FROM sys_user t WHERE t.account = ?"
	err := sqlDb.QueryRow(sqlStr, account).Scan(&count)

	if count <= 0 || err != nil {
		sqlResponse.Code = http.StatusBadRequest
		sqlResponse.Message = "删除的数据不存在"
		sqlResponse.Data = "error"
		c.JSON(http.StatusOK, sqlResponse)
		return
	}

	//删除
	delStr := "delete from sys_user t where t.account = ?"
	result, err := sqlDb.Exec(delStr, account)
	if err != nil {
		fmt.Printf("delete failed, err:%v\n", err)
		sqlResponse.Code = http.StatusBadRequest
		sqlResponse.Message = "删除失败"
		sqlResponse.Data = "error"
		c.JSON(http.StatusOK, sqlResponse)
		return
	}
	sqlResponse.Code = http.StatusOK
	sqlResponse.Message = "删除成功"
	sqlResponse.Data = "OK"
	c.JSON(http.StatusOK, sqlResponse)
	// 打印返回主键ID
	fmt.Println(result.LastInsertId())
}

/*
修改
*/
func updateData(c *gin.Context) {
	var u SqlUser
	// Bind() 通常用于从 HTTP 请求中解析请求体（例如 JSON 或 form 数据）并将其绑定到一个结构体对象上。
	err := c.Bind(&u)
	if err != nil {
		sqlResponse.Code = http.StatusBadRequest
		sqlResponse.Message = "参数错误"
		sqlResponse.Data = "error"
		// c.JSON() 方法将 sqlResponse 对象序列化成 JSON 格式
		//并设置响应的 Content-Type 为 "application/json"
		//同时会将 HTTP 状态码设置为 http.StatusOK，表示请求成功
		c.JSON(http.StatusOK, sqlResponse)
		return
	}
	sqlStr := "update sys_user t set t.account = ? where t.id = ?"
	ret, err := sqlDb.Exec(sqlStr, u.Account, u.Id)
	if err != nil {
		fmt.Printf("update failed, err:%v\n", err)
		sqlResponse.Code = http.StatusBadRequest
		sqlResponse.Message = "更新失败"
		sqlResponse.Data = "error"
		c.JSON(http.StatusOK, sqlResponse)
		return
	}
	sqlResponse.Code = http.StatusOK
	sqlResponse.Message = "更新成功"
	sqlResponse.Data = "OK"
	c.JSON(http.StatusOK, sqlResponse)
	fmt.Println(ret.LastInsertId()) //打印结果
}

/*
批量查询
*/
func getMulData(c *gin.Context) {
	// 用于获取 HTTP 请求 URL 中查询参数
	account := c.Query("id")
	sqlStr := "select t.id, t.account from sys_user t where t.id = ?"
	rows, err := sqlDb.Query(sqlStr, account)
	if err != nil {
		sqlResponse.Code = http.StatusBadRequest
		sqlResponse.Message = "查询错误"
		sqlResponse.Data = "error"
		c.JSON(http.StatusOK, sqlResponse)
		return
	}

	// 函数结束前关闭结果集
	defer rows.Close()

	resUser := make([]SqlUser, 0)
	// 遍历每一行存储到切片
	for rows.Next() {
		var userTemp SqlUser
		rows.Scan(&userTemp.Id, &userTemp.Account)
		// 追加到切片
		resUser = append(resUser, userTemp)
	}

	sqlResponse.Code = http.StatusOK
	sqlResponse.Message = "读取成功"
	sqlResponse.Data = resUser
	c.JSON(http.StatusOK, sqlResponse)
}

func getData(c *gin.Context) {
	id := c.Query("id")
	sqlStr := "select t.id, t.account from sys_user t where t.id = ?"
	var u SqlUser
	// Scan()通过顺序做绑定
	err := sqlDb.QueryRow(sqlStr, id).Scan(&u.Id, &u.Account)
	if err != nil {
		sqlResponse.Code = http.StatusBadRequest
		sqlResponse.Message = "查询错误"
		sqlResponse.Data = "error"
		c.JSON(http.StatusOK, sqlResponse)
		return
	}
	sqlResponse.Code = http.StatusOK
	sqlResponse.Message = "读取成功"
	sqlResponse.Data = u
	c.JSON(http.StatusOK, sqlResponse)
}

func insertData(c *gin.Context) {
	var u SqlUser
	err := c.Bind(&u)
	if err != nil {
		sqlResponse.Code = http.StatusBadRequest
		sqlResponse.Message = "参数错误"
		sqlResponse.Data = "error"
		c.JSON(http.StatusOK, sqlResponse)
		return
	}
	sqlStr := "insert into sys_user (account) values (?)"
	ret, err := sqlDb.Exec(sqlStr, u.Account)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		sqlResponse.Code = http.StatusBadRequest
		sqlResponse.Message = "写入失败"
		sqlResponse.Data = "error"
		c.JSON(http.StatusOK, sqlResponse)
		return
	}
	sqlResponse.Code = http.StatusOK
	sqlResponse.Message = "写入成功"
	sqlResponse.Data = "OK"
	c.JSON(http.StatusOK, sqlResponse)
	fmt.Println(ret.LastInsertId()) //打印结果

}
