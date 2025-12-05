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
	UserTicketLogStateEnabled  = 1 // 已启用
	UserTicketLogStateDisabled = 0 // 已禁用
)

type UserTicketLogDAO dbs.DAO

func NewUserTicketLogDAO() *UserTicketLogDAO {
	return dbs.NewDAO(&UserTicketLogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserTicketLogs",
			Model:  new(UserTicketLog),
			PkName: "id",
		},
	}).(*UserTicketLogDAO)
}

var SharedUserTicketLogDAO *UserTicketLogDAO

func init() {
	dbs.OnReady(func() {
		SharedUserTicketLogDAO = NewUserTicketLogDAO()
	})
}

// EnableUserTicketLog 启用
func (this *UserTicketLogDAO) EnableUserTicketLog(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserTicketLogStateEnabled).
		Update()
	return err
}

// DisableUserTicketLog 禁用
func (this *UserTicketLogDAO) DisableUserTicketLog(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserTicketLogStateDisabled).
		Update()
	return err
}

// FindEnabledUserTicketLog 查找启用中的日志
func (this *UserTicketLogDAO) FindEnabledUserTicketLog(tx *dbs.Tx, id int64) (*UserTicketLog, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(UserTicketLogStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserTicketLog), err
}

// CreateUserTicketLog 创建工单日志
func (this *UserTicketLogDAO) CreateUserTicketLog(tx *dbs.Tx, ticketId int64, adminId int64, userId int64, status userconfigs.UserTicketStatus, comment string, isReadonly bool) (int64, error) {
	if ticketId <= 0 {
		return 0, errors.New("invalid ticketId")
	}

	var now = time.Now().Unix()
	var op = NewUserTicketLogOperator()
	op.TicketId = ticketId
	op.AdminId = adminId
	op.UserId = userId
	if len(status) > 0 {
		op.Status = status
	}
	op.Comment = utils.LimitString(comment, 4096)
	op.CreatedAt = now
	op.IsReadonly = isReadonly
	op.State = UserTicketLogStateEnabled

	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}

	// 更新时间与状态
	_ = SharedUserTicketDAO.UpdateUserTicketLastLogAt(tx, ticketId, now)
	if len(status) > 0 {
		_ = SharedUserTicketDAO.UpdateUserTicketStatus(tx, ticketId, status)
	}

	return types.Int64(op.Id), nil
}

// DeleteUserTicketLog 删除工单日志
func (this *UserTicketLogDAO) DeleteUserTicketLog(tx *dbs.Tx, logId int64) error {
	if logId <= 0 {
		return errors.New("invalid userTicketLogId")
	}
	_, err := this.Query(tx).
		Pk(logId).
		Set("state", UserTicketLogStateDisabled).
		Update()
	return err
}

// CountUserTicketLogs 统计日志数量
func (this *UserTicketLogDAO) CountUserTicketLogs(tx *dbs.Tx, ticketId int64) (int64, error) {
	if ticketId <= 0 {
		return 0, nil
	}
	return this.Query(tx).
		Attr("ticketId", ticketId).
		State(UserTicketLogStateEnabled).
		Count()
}

// ListUserTicketLogs 列出日志
func (this *UserTicketLogDAO) ListUserTicketLogs(tx *dbs.Tx, ticketId int64, offset int64, size int64) (result []*UserTicketLog, err error) {
	_, err = this.Query(tx).
		Attr("ticketId", ticketId).
		State(UserTicketLogStateEnabled).
		Offset(offset).
		Limit(size).
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// FindLatestUserTicketLog 查找最新一条日志
func (this *UserTicketLogDAO) FindLatestUserTicketLog(tx *dbs.Tx, ticketId int64) (*UserTicketLog, error) {
	one, err := this.Query(tx).
		Attr("ticketId", ticketId).
		State(UserTicketLogStateEnabled).
		DescPk().
		Find()
	if one == nil {
		return nil, err
	}
	return one.(*UserTicketLog), err
}

// CheckUserTicketLog 检查日志是否属于指定工单
func (this *UserTicketLogDAO) CheckUserTicketLog(tx *dbs.Tx, ticketId int64, logId int64) (bool, error) {
	if ticketId <= 0 || logId <= 0 {
		return false, nil
	}
	return this.Query(tx).
		Pk(logId).
		Attr("ticketId", ticketId).
		State(UserTicketLogStateEnabled).
		Exist()
}
