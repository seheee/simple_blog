package controllers

import (
	"fmt"
	"net/http"
	"simpleblog/models"
	"strconv"

	"github.com/labstack/echo"
)

type PostController struct {
}

// 이 함수들은 핸들러임

// main함수에서 controllers.PostController{}.Init(e.Group("/posts"))를 호출함
// -> post model을 처리하기 위한 post controller의 핸들러 함수들을 각각의 http요청에 등록하는것
// 여기서 e.Group은 url요청에서 앞부분이 /posts로 시작하는 요청들을 그룹으로 나누겠다는것
func (c PostController) Init(g *echo.Group) {
	g.GET("/", c.Index)
	g.GET("/:id", c.GetById)
	g.GET("/create", func(c echo.Context) error {
		return c.Render(http.StatusOK, "Create.html", nil)
	})
	// ip:port/posts/create의 요청이 오면 PostController.Create를 통해서 처리하겠다는 의미
	g.POST("/create", c.Create)
	g.GET("/:id/delete", c.Delete)
}

// 클라이언트가 모든 post목록을 보여달라는 요청 보내면 postController.Index메소드 방식으로 처리하겠다는 의미
func (PostController) Index(c echo.Context) error {
	// post모델의 index메소드 이용해서 post정보 얻어옴
	// controllers.PostController.Index의 전달인자는 echo.Context
	// models.Post.Index의 전달인자는 context.Context
	// -> echo.Context.Request().Context()를 통해서 context.Context얻음
	ps, err := models.Post{}.Index(c.Request().Context())
	if err != nil {
		c.String(http.StatusOK, "read DB fail")
	}

	// db로부터 얻어온 정보를 index.html에 렌더링하여 요청 처리
	// index.html에서 {{~~}}부분(template)에 ps를 채워넣어서 클라이언트에게 전달
	return c.Render(http.StatusOK, "Index.html", ps)
}

func (PostController) GetById(c echo.Context) error {
	// url요청에서 id정보를 c.Param(key)를 이용해서 얻을 수 있다.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusOK, fmt.Sprintf("%s is not integer", c.Param("id")))
	}
	// Post모델의 GetById 메소드를 이용해서 원하는 정보 얻어옴
	p, err := models.Post{}.GetById(c.Request().Context(), id)
	if err != nil {
		c.String(http.StatusOK, fmt.Sprintf("%d post isn't exist", id))
	}
	// Post.html에 렌더링해서 요청 처리
	return c.Render(http.StatusOK, "Post.html", p)
}

// post 요청을 처리하는 핸들러
// 프론트에서 title, body의 내용을 form형식을 통해서 보내도록 설정함
func (PostController) Create(c echo.Context) error {
	p := new(models.Post)
	// echo.Context.FormValue를 통해 form내용을 읽을 수 있음
	p.Title = c.FormValue("title")
	p.Body = c.FormValue("body")

	// post모델의 create메소드를 이용해서 새로운 post tuple을 db에 저장
	// 이때 Post.CreatedAt, UpdatedAt은 xorm이 알아서 갱신
	err := p.Create(c.Request().Context())
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	// Echo.Context.Redirect를 통해서 클라이언트에게 경로 재지정
	// -> ip:port/posts/create에서 새로운 post를 생성하라는 요청 받으면 요청을 처리하고
	//		ip:/por/posts/로 이동시키는 작업을 하는것
	return c.Redirect(http.StatusMovedPermanently, "/posts/")
}

func (PostController) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusOK, fmt.Sprintf("%s is not integer", c.Param("id")))
	}

	models.Post{}.Delete(c.Request().Context(), id)
	return c.Redirect(http.StatusMovedPermanently, "/posts/")
}
