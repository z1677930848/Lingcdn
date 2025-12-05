package services

import (
	"fmt"

	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

func buildPBUserTicketCategory(tx *dbs.Tx, categoryId int64, cacheMap *utils.CacheMap) (*pb.UserTicketCategory, error) {
	if categoryId <= 0 {
		return nil, nil
	}
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = "userTicketCategory:" + types.String(categoryId)
	if cacheValue, ok := cacheMap.Get(cacheKey); ok {
		if cacheValue == nil {
			return nil, nil
		}
		return cacheValue.(*pb.UserTicketCategory), nil
	}

	category, err := models.SharedUserTicketCategoryDAO.FindEnabledUserTicketCategory(tx, categoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		cacheMap.Put(cacheKey, nil)
		return nil, nil
	}
	var pbCategory = &pb.UserTicketCategory{
		Id:   int64(category.Id),
		Name: category.Name,
		IsOn: category.IsOn == 1,
	}
	cacheMap.Put(cacheKey, pbCategory)
	return pbCategory, nil
}

func buildPBAdmin(tx *dbs.Tx, adminId int64, cacheMap *utils.CacheMap) (*pb.Admin, error) {
	if adminId <= 0 {
		return nil, nil
	}
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	cacheKey := fmt.Sprintf("admin:%d", adminId)
	if cacheValue, ok := cacheMap.Get(cacheKey); ok {
		if cacheValue == nil {
			return nil, nil
		}
		return cacheValue.(*pb.Admin), nil
	}

	admin, err := models.SharedAdminDAO.FindBasicAdmin(tx, adminId)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		cacheMap.Put(cacheKey, nil)
		return nil, nil
	}

	pbAdmin := &pb.Admin{
		Id:       int64(admin.Id),
		Username: admin.Username,
		Fullname: admin.Fullname,
		IsOn:     admin.IsOn,
		IsSuper:  admin.IsSuper,
		CanLogin: admin.CanLogin,
	}
	cacheMap.Put(cacheKey, pbAdmin)
	return pbAdmin, nil
}

func buildPBUser(tx *dbs.Tx, userId int64, cacheMap *utils.CacheMap) (*pb.User, error) {
	if userId <= 0 {
		return nil, nil
	}
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	cacheKey := fmt.Sprintf("user:%d", userId)
	if cacheValue, ok := cacheMap.Get(cacheKey); ok {
		if cacheValue == nil {
			return nil, nil
		}
		return cacheValue.(*pb.User), nil
	}

	user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		cacheMap.Put(cacheKey, nil)
		return nil, nil
	}

	pbUser := &pb.User{
		Id:       int64(user.Id),
		Username: user.Username,
		Fullname: user.Fullname,
		IsOn:     user.IsOn,
	}
	cacheMap.Put(cacheKey, pbUser)
	return pbUser, nil
}

func buildPBUserTicketLog(tx *dbs.Tx, log *models.UserTicketLog, cacheMap *utils.CacheMap) (*pb.UserTicketLog, error) {
	if log == nil {
		return nil, nil
	}
	admin, err := buildPBAdmin(tx, int64(log.AdminId), cacheMap)
	if err != nil {
		return nil, err
	}
	user, err := buildPBUser(tx, int64(log.UserId), cacheMap)
	if err != nil {
		return nil, err
	}
	return &pb.UserTicketLog{
		Id:         int64(log.Id),
		AdminId:    int64(log.AdminId),
		UserId:     int64(log.UserId),
		TicketId:   int64(log.TicketId),
		Status:     log.Status,
		Comment:    log.Comment,
		CreatedAt:  int64(log.CreatedAt),
		IsReadonly: log.IsReadonly == 1,
		Admin:      admin,
		User:       user,
	}, nil
}

func buildPBUserTicket(tx *dbs.Tx, ticket *models.UserTicket, cacheMap *utils.CacheMap) (*pb.UserTicket, error) {
	if ticket == nil {
		return nil, nil
	}
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}

	category, err := buildPBUserTicketCategory(tx, int64(ticket.CategoryId), cacheMap)
	if err != nil {
		return nil, err
	}
	user, err := buildPBUser(tx, int64(ticket.UserId), cacheMap)
	if err != nil {
		return nil, err
	}
	var latestLog *pb.UserTicketLog
	log, err := models.SharedUserTicketLogDAO.FindLatestUserTicketLog(tx, int64(ticket.Id))
	if err != nil {
		return nil, err
	}
	if log != nil {
		latestLog, err = buildPBUserTicketLog(tx, log, cacheMap)
		if err != nil {
			return nil, err
		}
	}

	return &pb.UserTicket{
		Id:                  int64(ticket.Id),
		CategoryId:          int64(ticket.CategoryId),
		UserId:              int64(ticket.UserId),
		Subject:             ticket.Subject,
		Body:                ticket.Body,
		Status:              ticket.Status,
		CreatedAt:           int64(ticket.CreatedAt),
		LastLogAt:           int64(ticket.LastLogAt),
		UserTicketCategory:  category,
		User:                user,
		LatestUserTicketLog: latestLog,
	}, nil
}
