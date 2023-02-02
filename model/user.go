package model

type User struct {
	ID    uint64 `sql:"primary_key" json:"id"`
	Name  string `sql:"not null;type:varchar(255);" json:"name"`
	Title string `sql:"not null;type:varchar(255);" json:"title"`
}
