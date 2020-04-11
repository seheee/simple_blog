package models

import (
	"context"
	"time"

	"github.com/go-xorm/xorm"
)

type Post struct {
	Id        int64
	Title     string
	Body      string `xorm:"text"`
	Comments  []Comment
	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
}

// DB에 등록된 모든 Post를 조회해서 슬라이스에 저장해서 리턴해줌
func (Post) Index(c context.Context) ([]Post, error) {
	var ps []Post

	// 핸들러가 호출되는 과정에서 미들웨어에서 db session을 context에 등록함
	db := c.Value("DB").(*xorm.Session) // "DB"이름으로 등록한 값을 *xorm.Session타입으로 형변환
	rows, err := db.Rows(&Post{}) // 구조체에 해당하는 table을 db에서 알아서 검색해서 모든 행 반환
	if err != nil {
		return nil, err
	}

	for rows.Next() { // 행을 순회, 마지막 행이라면 false 리턴
		p := Post{}
		err = rows.Scan(&p) // 해당 구조체에 db로부터 얻어온 tuple정보를 저장
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}

	return ps, nil
}

// post table에서 primary key인 id값을 이용하여 특정 post검색, 구조체에 저장하여 리턴
func (Post) GetById(c context.Context, id int) (*Post, error) {
	p := new(Post)

	db := c.Value("DB").(*xorm.Session)
	ok, err := db.Id(id).Get(p) // 구조체포인터(p)에 해당하는 table에서 id를 이용해서 특정 tuple 검색
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}
	// Comment 모델이 지원하는 GetById 메소드를 호출하여 post에 해당하는 comment 검색
	p.Comments, err = Comment{}.GetById(c, p.Id)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// 해당 구조체를 db(post table)에 추가
func (p *Post) Create(c context.Context) error {
	db := c.Value("DB").(*xorm.Session)
	_, err := db.Insert(p)
	return err
}

// 해당 구조체를 db(post table)에서 삭제
func (Post) Delete(c context.Context, id int) error {
	db := c.Value("DB").(*xorm.Session)
	_, err := db.ID(id).Delete(&Post{}) // // 구조체포인터(p)에 해당하는 table에서 id를 이용해서 특정 tuple 삭제
	return err
}
