package services

import (
	"context"

	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// UserTicketCategoryService 工单分类相关服务
type UserTicketCategoryService struct {
	BaseService
}

// CreateUserTicketCategory 创建分类
func (this *UserTicketCategoryService) CreateUserTicketCategory(ctx context.Context, req *pb.CreateUserTicketCategoryRequest) (*pb.CreateUserTicketCategoryResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	categoryId, err := models.SharedUserTicketCategoryDAO.CreateUserTicketCategory(tx, req.Name)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserTicketCategoryResponse{UserTicketCategoryId: categoryId}, nil
}

// UpdateUserTicketCategory 修改分类
func (this *UserTicketCategoryService) UpdateUserTicketCategory(ctx context.Context, req *pb.UpdateUserTicketCategoryRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedUserTicketCategoryDAO.UpdateUserTicketCategory(tx, req.UserTicketCategoryId, req.Name, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteUserTicketCategory 删除分类
func (this *UserTicketCategoryService) DeleteUserTicketCategory(ctx context.Context, req *pb.DeleteUserTicketCategoryRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedUserTicketCategoryDAO.DisableUserTicketCategory(tx, req.UserTicketCategoryId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllUserTicketCategories 查询所有分类
func (this *UserTicketCategoryService) FindAllUserTicketCategories(ctx context.Context, req *pb.FindAllUserTicketCategoriesRequest) (*pb.FindAllUserTicketCategoriesResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	categories, err := models.SharedUserTicketCategoryDAO.FindAllUserTicketCategories(tx)
	if err != nil {
		return nil, err
	}

	result := make([]*pb.UserTicketCategory, 0, len(categories))
	for _, category := range categories {
		result = append(result, &pb.UserTicketCategory{
			Id:   int64(category.Id),
			Name: category.Name,
			IsOn: category.IsOn == 1,
		})
	}

	return &pb.FindAllUserTicketCategoriesResponse{UserTicketCategories: result}, nil
}

// FindAllAvailableUserTicketCategories 查询启用中的分类
func (this *UserTicketCategoryService) FindAllAvailableUserTicketCategories(ctx context.Context, req *pb.FindAllAvailableUserTicketCategoriesRequest) (*pb.FindAllAvailableUserTicketCategoriesResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	categories, err := models.SharedUserTicketCategoryDAO.FindAllAvailableUserTicketCategories(tx)
	if err != nil {
		return nil, err
	}

	result := make([]*pb.UserTicketCategory, 0, len(categories))
	for _, category := range categories {
		result = append(result, &pb.UserTicketCategory{
			Id:   int64(category.Id),
			Name: category.Name,
			IsOn: category.IsOn == 1,
		})
	}

	return &pb.FindAllAvailableUserTicketCategoriesResponse{UserTicketCategories: result}, nil
}

// FindUserTicketCategory 查询单个分类
func (this *UserTicketCategoryService) FindUserTicketCategory(ctx context.Context, req *pb.FindUserTicketCategoryRequest) (*pb.FindUserTicketCategoryResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	category, err := models.SharedUserTicketCategoryDAO.FindEnabledUserTicketCategory(tx, req.UserTicketCategoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return &pb.FindUserTicketCategoryResponse{UserTicketCategory: nil}, nil
	}

	return &pb.FindUserTicketCategoryResponse{
		UserTicketCategory: &pb.UserTicketCategory{
			Id:   int64(category.Id),
			Name: category.Name,
			IsOn: category.IsOn == 1,
		},
	}, nil
}
