package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	UserTicketCategoryStateEnabled  = 1 // 已启用
	UserTicketCategoryStateDisabled = 0 // 已禁用
)

type UserTicketCategoryDAO dbs.DAO

func NewUserTicketCategoryDAO() *UserTicketCategoryDAO {
	return dbs.NewDAO(&UserTicketCategoryDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserTicketCategories",
			Model:  new(UserTicketCategory),
			PkName: "id",
		},
	}).(*UserTicketCategoryDAO)
}

var SharedUserTicketCategoryDAO *UserTicketCategoryDAO

func init() {
	dbs.OnReady(func() {
		SharedUserTicketCategoryDAO = NewUserTicketCategoryDAO()
	})
}

// EnableUserTicketCategory 启用分类
func (this *UserTicketCategoryDAO) EnableUserTicketCategory(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserTicketCategoryStateEnabled).
		Update()
	return err
}

// DisableUserTicketCategory 禁用分类
func (this *UserTicketCategoryDAO) DisableUserTicketCategory(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserTicketCategoryStateDisabled).
		Update()
	return err
}

// FindEnabledUserTicketCategory 查找启用中的分类
func (this *UserTicketCategoryDAO) FindEnabledUserTicketCategory(tx *dbs.Tx, id int64) (*UserTicketCategory, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(UserTicketCategoryStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserTicketCategory), err
}

// CreateUserTicketCategory 创建分类
func (this *UserTicketCategoryDAO) CreateUserTicketCategory(tx *dbs.Tx, name string) (int64, error) {
	if len(name) == 0 {
		return 0, errors.New("category name should not be empty")
	}

	var op = NewUserTicketCategoryOperator()
	op.Name = name
	op.IsOn = true
	op.State = UserTicketCategoryStateEnabled

	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateUserTicketCategory 修改分类
func (this *UserTicketCategoryDAO) UpdateUserTicketCategory(tx *dbs.Tx, categoryId int64, name string, isOn bool) error {
	if categoryId <= 0 {
		return errors.New("invalid userTicketCategoryId")
	}

	var op = NewUserTicketCategoryOperator()
	op.Id = categoryId
	if len(name) > 0 {
		op.Name = name
	}
	op.IsOn = isOn

	return this.Save(tx, op)
}

// FindAllUserTicketCategories 查找全部分类
func (this *UserTicketCategoryDAO) FindAllUserTicketCategories(tx *dbs.Tx) (result []*UserTicketCategory, err error) {
	_, err = this.Query(tx).
		State(UserTicketCategoryStateEnabled).
		Asc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllAvailableUserTicketCategories 查找启用中的分类
func (this *UserTicketCategoryDAO) FindAllAvailableUserTicketCategories(tx *dbs.Tx) (result []*UserTicketCategory, err error) {
	_, err = this.Query(tx).
		State(UserTicketCategoryStateEnabled).
		Attr("isOn", true).
		Asc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}
