package repo

import (
	"github.com/sakiib/integrate-vault-with-micriservices-in-k8s/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Get(id int) (*model.User, error)
	Update(*model.User) error
}

type PostgresUser struct {
	db *gorm.DB
}

func NewUserPostgresRepository(db *gorm.DB) UserRepository {
	return &PostgresUser{db: db}
}

// Create ...
func (p *PostgresUser) Get(id int) (*model.User, error) {
	user := model.User{}

	qry := p.db.
		Table("users").
		Select("*").
		Where("id = ?", id).
		Limit(1)

	if err := qry.First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// Update ...
func (p *PostgresUser) Update(m *model.User) error {
	return p.db.Model(m).UpdateColumns(m).Error
}
