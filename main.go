package main

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"simpleblog/controllers"
	"simpleblog/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	driver  = "mysql"
	connect = "root:sehee@tcp(172.16.11.203:3306)/test_db?charset=utf8"
	//driver  = "sqlite3"
	//connect = "./blog.db"
)

func main() {
	db, err := xorm.NewEngine(driver, connect) // xorm을 통해 db와 연결
	if err != nil {
		panic(err)
	}
	defer db.Close() // main함수 종료되면 db연결 종료
	// xorm watches tables and indexes and sync schema:
	// sync를 통해 post, comment 테이블 생성, 수정
	err = db.Sync(new(models.Post))
	err = db.Sync(new(models.Comment))

	e := echo.New() // echo 객체 생성
	// 미들웨어 등록
	// 미들웨어 -> 루트레벨, 그룹레벨, 라우트레벨
	// 루트레벨 미들웨어 중 라우트이전 미들웨어는 echo.Pre메소드로, 라우트이후 미들웨어는 echo.Use로 설정
	// 그룹레벨 미들웨어는 특정 주소 하위패스에만 미들웨어로 추가로 적용하는 것
	e.Use(ContextDB(db))        // db에 접근하기 위해 따로 구현한 미들웨어
	e.Use(middleware.Logger())  // echo에서 지원해주는 것
	e.Use(middleware.Recover()) // echo에서 지원해주는 것

	// 구조체 생성하고 Init (핸들러 등록)
	controllers.PostController{}.Init(e.Group("/posts"))                // post기능 핸들러 등록
	controllers.CommentController{}.Init(e.Group("/posts/:id/comment")) // comment기능 핸들러 등록

	t := &Template{
		templates: template.Must(template.ParseGlob("./views/*.html")),
	}
	e.Renderer = t

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Home")
	})

	e.Logger.Fatal(e.Start(":1323"))
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func ContextDB(db *xorm.Engine) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session := db.NewSession()
			defer session.Close()

			req := c.Request()
			c.SetRequest(req.WithContext(context.WithValue(req.Context(), "DB", session)))

			switch req.Method {
			case "POST", "PUT", "DELETE":
				if err := session.Begin(); err != nil {
					return echo.NewHTTPError(500, err.Error())
				}
				if err := next(c); err != nil {
					session.Rollback()
					return echo.NewHTTPError(500, err.Error())
				}
				if c.Response().Status >= 500 {
					session.Rollback()
					return nil
				}
				if err := session.Commit(); err != nil {
					return echo.NewHTTPError(500, err.Error())
				}
			default:
				if err := next(c); err != nil {
					return echo.NewHTTPError(500, err.Error())
				}
			}

			return nil
		}
	}
}
