package models

import (
	"time"

	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	UserTicketStateEnabled  = 1 // 已启用
	UserTicketStateDisabled = 0 // 已禁用
)

type UserTicketDAO dbs.DAO

func NewUserTicketDAO() *UserTicketDAO {
	return dbs.NewDAO(&UserTicketDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserTickets",
			Model:  new(UserTicket),
			PkName: "id",
		},
	}).(*UserTicketDAO)
}

var SharedUserTicketDAO *UserTicketDAO

func init() {
	dbs.OnReady(func() {
		SharedUserTicketDAO = NewUserTicketDAO()
	})
}

// EnableUserTicket 启用
func (this *UserTicketDAO) EnableUserTicket(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserTicketStateEnabled).
		Update()
	return err
}

// DisableUserTicket 禁用
func (this *UserTicketDAO) DisableUserTicket(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserTicketStateDisabled).
		Update()
	return err
}

// FindEnabledUserTicket 查找启用中的工单
func (this *UserTicketDAO) FindEnabledUserTicket(tx *dbs.Tx, id int64) (*UserTicket, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(UserTicketStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserTicket), err
}

// CreateUserTicket 创建工单
func (this *UserTicketDAO) CreateUserTicket(tx *dbs.Tx, userId int64, categoryId int64, subject string, body string) (int64, error) {
	if userId <= 0 {
		return 0, errors.New("invalid userId")
	}

	var now = time.Now().Unix()
	var op = NewUserTicketOperator()
	op.UserId = userId
	op.CategoryId = categoryId
	op.Subject = utils.LimitString(subject, 255)
	op.Body = utils.LimitString(body, 4096)
	op.Status = userconfigs.UserTicketStatusNone
	op.CreatedAt = now
	op.LastLogAt = now
	op.State = UserTicketStateEnabled

	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateUserTicket 修改工单
func (this *UserTicketDAO) UpdateUserTicket(tx *dbs.Tx, ticketId int64, categoryId int64, subject string, body string) error {
	if ticketId <= 0 {
		return errors.New("invalid userTicketId")
	}

	var op = NewUserTicketOperator()
	op.Id = ticketId
	if categoryId > 0 {
		op.CategoryId = categoryId
	}
	if len(subject) > 0 {
		op.Subject = utils.LimitString(subject, 255)
	}
	if len(body) > 0 {
		op.Body = utils.LimitString(body, 4096)
	}

	return this.Save(tx, op)
}

// DeleteUserTicket 删除工单
func (this *UserTicketDAO) DeleteUserTicket(tx *dbs.Tx, ticketId int64) error {
	if ticketId <= 0 {
		return errors.New("invalid userTicketId")
	}
	_, err := this.Query(tx).
		Pk(ticketId).
		Set("state", UserTicketStateDisabled).
		Update()
	return err
}

// CountUserTickets 统计工单数量
func (this *UserTicketDAO) CountUserTickets(tx *dbs.Tx, userId int64, categoryId int64, status string) (int64, error) {
	query := this.Query(tx).
		State(UserTicketStateEnabled)
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if categoryId > 0 {
		query.Attr("categoryId", categoryId)
	}
	if len(status) > 0 {
		query.Attr("status", status)
	}
	return query.Count()
}

// ListUserTickets 列出工单
func (this *UserTicketDAO) ListUserTickets(tx *dbs.Tx, userId int64, categoryId int64, status string, offset int64, size int64) (result []*UserTicket, err error) {
	query := this.Query(tx).
		State(UserTicketStateEnabled)
	if userId > 0 {
		query.Attr("userId", userId)
	}
	if categoryId > 0 {
		query.Attr("categoryId", categoryId)
	}
	if len(status) > 0 {
		query.Attr("status", status)
	}

	_, err = query.
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CheckUserTicket 检查工单是否属于用户
func (this *UserTicketDAO) CheckUserTicket(tx *dbs.Tx, userId int64, ticketId int64) (bool, error) {
	if userId <= 0 || ticketId <= 0 {
		return false, nil
	}
	return this.Query(tx).
		Pk(ticketId).
		Attr("userId", userId).
		State(UserTicketStateEnabled).
		Exist()
}

// UpdateUserTicketStatus 更新工单状态
func (this *UserTicketDAO) UpdateUserTicketStatus(tx *dbs.Tx, ticketId int64, status userconfigs.UserTicketStatus) error {
	if ticketId <= 0 {
		return errors.New("invalid userTicketId")
	}
	if len(status) == 0 {
		return errors.New("invalid status")
	}
	var op = NewUserTicketOperator()
	op.Id = ticketId
	op.Status = status
	return this.Save(tx, op)
}

// UpdateUserTicketLastLogAt 更新工单最后日志时间
func (this *UserTicketDAO) UpdateUserTicketLastLogAt(tx *dbs.Tx, ticketId int64, timestamp int64) error {
	if ticketId <= 0 {
		return errors.New("invalid userTicketId")
	}
	var op = NewUserTicketOperator()
	op.Id = ticketId
	op.LastLogAt = timestamp
	return this.Save(tx, op)
}

// FindUserTicketUserId 获取工单所属用户
func (this *UserTicketDAO) FindUserTicketUserId(tx *dbs.Tx, ticketId int64) (int64, error) {
	return this.Query(tx).
		Pk(ticketId).
		State(UserTicketStateEnabled).
		Result("userId").
		FindInt64Col(0)
}
