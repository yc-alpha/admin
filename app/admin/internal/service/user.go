package service

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"entgo.io/ent/dialect/sql"
	v1 "github.com/yc-alpha/admin/api/admin/v1"
	"github.com/yc-alpha/admin/app/admin/internal/data/ent"
	"github.com/yc-alpha/admin/app/admin/internal/data/ent/sysuser"
	"github.com/yc-alpha/config"
	"github.com/yc-alpha/variant"
)

type UserService struct {
	v1.UnimplementedUserServiceServer
	client *ent.Client
}

func NewUserService(client *ent.Client) *UserService {
	return &UserService{
		client: client,
	}
}

func convertUserAccountToProto(account *ent.SysUserAccount) *v1.UserAccount {
	target := &v1.UserAccount{
		UserId:     strconv.FormatInt(account.UserID, 10),
		Platform:   account.Platform,
		Identifier: account.Identifier,
		Name:       variant.New(account.Name).ToString(),
		CreatedAt:  account.CreatedAt.Format(time.DateTime),
		UpdatedAt:  account.UpdatedAt.Format(time.DateTime),
	}
	return target
}

func convertSimpleUserToProto(user *ent.SysUser) *v1.SimpleUser {
	return &v1.SimpleUser{
		Id:        strconv.FormatInt(user.ID, 10),
		Username:  user.Username,
		Email:     variant.New(user.Email).ToString(),
		Phone:     variant.New(user.Phone).ToString(),
		Fullname:  variant.New(user.FullName).ToString(),
		Avatar:    variant.New(user.Avatar).ToString(),
		Status:    v1.UserStatus(v1.UserStatus_value[user.Status.String()]),
		Gender:    v1.Gender(v1.Gender_value[user.Gender.String()]),
		Timezone:  user.Timezone,
		Language:  user.Language,
		CreatedBy: variant.New(user.CreatedBy).ToString(),
		UpdatedBy: variant.New(user.UpdatedBy).ToString(),
		CreatedAt: user.CreatedAt.Format(time.DateTime),
		UpdatedAt: user.UpdatedAt.Format(time.DateTime),
	}
}

func convertUserToProto(user *ent.SysUser, accounts ...*ent.SysUserAccount) *v1.User {
	target := &v1.User{
		Id:        strconv.FormatInt(user.ID, 10),
		Username:  user.Username,
		Email:     variant.New(user.Email).ToString(),
		Phone:     variant.New(user.Phone).ToString(),
		Fullname:  variant.New(user.FullName).ToString(),
		Avatar:    variant.New(user.Avatar).ToString(),
		Status:    v1.UserStatus(v1.UserStatus_value[user.Status.String()]),
		Gender:    v1.Gender(v1.Gender_value[user.Gender.String()]),
		Timezone:  user.Timezone,
		Language:  user.Language,
		CreatedBy: variant.New(user.CreatedBy).ToString(),
		UpdatedBy: variant.New(user.UpdatedBy).ToString(),
		CreatedAt: user.CreatedAt.Format(time.DateTime),
		UpdatedAt: user.UpdatedAt.Format(time.DateTime),
	}

	for _, account := range accounts {
		target.UserAccounts = append(target.UserAccounts, convertUserAccountToProto(account))
	}
	return target
}

// CreateUser creates a new user in the system.
func (s *UserService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	// Validate input parameters.
	if req.GetUsername() == "" {
		return &v1.CreateUserResponse{Result: false, Code: 500, User: nil, Msg: "username is required"}, nil
	}
	if req.GetEmail() == "" && req.GetPhone() == "" {
		return &v1.CreateUserResponse{Result: false, Code: 500, User: nil, Msg: "user email or phone is required"}, nil
	}
	// Start a transaction.
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return &v1.CreateUserResponse{Result: false, Code: 500, User: nil, Msg: fmt.Errorf("starting a transaction: %w", err).Error()}, nil
	}
	defer tx.Rollback()

	creator := tx.SysUser.Create().
		SetUsername(req.Username).
		SetPassword(req.Password).
		SetEmail(req.Email).
		SetPhone(req.Phone).
		SetFullName(req.Fullname).
		SetStatus(sysuser.Status(req.Status.String())).
		SetGender(sysuser.Gender(req.Gender.String())).
		SetLanguage(req.Language).
		SetTimezone(req.Timezone)
	if config.GetBool("system.skip_activate", false) {
		creator.SetStatus(sysuser.StatusACTIVE)
	}
	user, err := creator.Save(ctx)
	if err != nil {
		return nil, err
	}

	var userAccounts []*ent.SysUserAccount
	if len(req.UserAccounts) > 0 {
		for _, account := range req.UserAccounts {
			accBuilder := tx.SysUserAccount.Create().
				SetUserID(user.ID).
				SetPlatform(account.Platform).
				SetIdentifier(account.Identifier)
			if account.Name != "" {
				accBuilder.SetName(account.Name)
			}
			acc, err := accBuilder.Save(ctx)
			if err != nil {
				return nil, err
			}
			userAccounts = append(userAccounts, acc)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &v1.CreateUserResponse{
		Result: true,
		Code:   200,
		User:   convertUserToProto(user, userAccounts...),
		Msg:    "success",
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
	userID, err := strconv.ParseInt(req.GetId(), 10, 64)
	if err != nil {
		return &v1.DeleteUserResponse{Result: false, Code: 500, Msg: "invalid user ID"}, nil
	}
	if err := s.client.SysUser.DeleteOneID(userID).Exec(ctx); err != nil {
		return &v1.DeleteUserResponse{Result: false, Code: 500, Msg: "failed to delete user: " + err.Error()}, nil
	}
	return &v1.DeleteUserResponse{Result: true, Code: 200, Msg: "user deleted successfully"}, nil
}

// UpdateUser updates an existing user in the system.
func (s *UserService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
	userID, err := strconv.ParseInt(req.GetId(), 10, 64)
	if err != nil {
		return &v1.UpdateUserResponse{Result: false, Code: 500, Msg: "invalid user ID"}, nil
	}
	updater := s.client.SysUser.UpdateOneID(userID).
		SetUsername(req.Username).
		SetEmail(req.Email).
		SetPhone(req.Phone).
		SetStatus(sysuser.Status(req.Status.String())).
		SetGender(sysuser.Gender(req.Gender.String())).
		SetLanguage(req.Language).
		SetTimezone(req.Timezone)

	err = updater.Exec(ctx)
	if err != nil {
		return &v1.UpdateUserResponse{Result: false, Code: 500, Msg: "failed to update user: " + err.Error(), User: nil}, nil
	}

	return &v1.UpdateUserResponse{
		Result: true,
		Code:   200,
		Msg:    "user updated successfully",
		User:   &v1.SimpleUser{},
	}, nil
}

func (s *UserService) UpdateUserAccounts(ctx context.Context, req *v1.UpdateUserAccountsRequest) (*v1.UpdateUserAccountsResponse, error) {
	userID, err := strconv.ParseInt(req.GetId(), 10, 64)
	if err != nil {
		return &v1.UpdateUserAccountsResponse{Result: false, Code: 500, Msg: "invalid user ID"}, nil
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return &v1.UpdateUserAccountsResponse{Result: false, Code: 500, Msg: "failed to start transaction"}, nil
	}
	defer tx.Rollback()

	// lock users
	user, err := tx.SysUser.Query().Where(sysuser.ID(userID)).WithAccounts().ForUpdate().Only(ctx)
	if err != nil {
		return &v1.UpdateUserAccountsResponse{Result: false, Code: 500, Msg: "user not found"}, nil
	}

	// Create old account mapping (using Platform+Account as a unique identifier)
	oldAccountMap := make(map[string]*ent.SysUserAccount)
	for _, acc := range user.Edges.Accounts {
		key := acc.Platform + "|" + acc.Identifier
		oldAccountMap[key] = acc
	}

	// Create a new account mapping
	newAccountMap := make(map[string]*ent.SysUserAccount)
	for _, acc := range req.UserAccounts {
		key := acc.Platform + "|" + acc.Identifier
		account := &ent.SysUserAccount{
			UserID:     userID,
			Platform:   acc.Platform,
			Identifier: acc.Identifier,
			Name:       &acc.Name,
		}
		newAccountMap[key] = account
	}

	// Find the account that needs to be deleted
	for key := range oldAccountMap {
		if _, exists := newAccountMap[key]; !exists {
			if err := tx.SysUserAccount.DeleteOneID(oldAccountMap[key].ID).Exec(ctx); err != nil {
				return &v1.UpdateUserAccountsResponse{Result: false, Code: 500, Msg: "failed to delete user account: " + err.Error()}, nil
			}
		}
	}

	// Update or create user accounts
	for _, newAcc := range newAccountMap {
		if err := tx.SysUserAccount.
			Create().
			SetPlatform(newAcc.Platform).
			SetIdentifier(newAcc.Identifier).
			SetName(*newAcc.Name).
			SetUserID(newAcc.UserID).
			OnConflict(
				sql.ConflictColumns("platform", "identifier"), // 唯一键
			).
			UpdateNewValues().
			Exec(ctx); err != nil {
			return &v1.UpdateUserAccountsResponse{Result: false, Code: 500, Msg: "failed to update or create user account: " + err.Error()}, nil
		}
	}

	if err = tx.Commit(); err != nil {
		return &v1.UpdateUserAccountsResponse{Result: false, Code: 500, Msg: "failed to commit transaction"}, nil
	}

	return &v1.UpdateUserAccountsResponse{Result: true, Code: 200, Msg: "user accounts updated successfully"}, nil
}

func (s *UserService) GetUserInfo(ctx context.Context, req *v1.GetUserInfoRequest) (*v1.GetUserInfoResponse, error) {
	if req.Id == "" && req.Username == "" && req.Email == "" && req.Phone == "" {
		return &v1.GetUserInfoResponse{Result: false, Code: 500, Msg: "user ID, username, email, or phone is required"}, nil
	}
	user, err := s.client.SysUser.Query().
		Where(sysuser.Or(
			sysuser.ID(variant.New(req.GetId()).ToInt64()),
			sysuser.Username(req.GetUsername()),
			sysuser.Email(req.GetEmail()),
			sysuser.Phone(req.GetPhone()),
		)).
		WithAccounts().
		Only(ctx)
	if err != nil {
		return &v1.GetUserInfoResponse{Result: false, Code: 500, Msg: "user not found"}, nil
	}

	return &v1.GetUserInfoResponse{
		Result: true,
		Code:   200,
		User:   convertUserToProto(user, user.Edges.Accounts...),
		Msg:    "user retrieved successfully",
	}, nil
}

// ListUsers retrieves a list of users based on the provided filters and pagination.
func (s *UserService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	q := s.client.SysUser.Query()

	if req.GetUsername() != "" {
		q.Where(sysuser.UsernameContains(req.GetUsername()))
	}
	if req.GetEmail() != "" {
		q.Where(sysuser.EmailContains(req.GetEmail()))
	}
	if req.GetPhone() != "" {
		q.Where(sysuser.PhoneContains(req.GetPhone()))
	}
	if req.GetFilter() != "" {
		q.Where(sysuser.Or(
			sysuser.UsernameContains(req.GetUsername()),
			sysuser.EmailContains(req.GetEmail()),
			sysuser.PhoneContains(req.GetPhone()),
		))
	}
	if len(req.GetStatus()) > 0 && len(req.GetStatus()) < 3 {
		s := []sysuser.Status{}
		for _, status := range req.GetStatus() {
			s = append(s, sysuser.Status(status.String()))
		}
		q.Where(sysuser.StatusIn(s...))
	}
	// 排序参数
	allowedOrderFields := []string{
		sysuser.FieldUsername,
		sysuser.FieldCreatedAt,
	}
	if slices.Contains(allowedOrderFields, req.GetOrder()) {
		if req.GetIsDesc() {
			q.Order(ent.Desc(req.GetOrder()))
		} else {
			q.Order(ent.Asc(req.GetOrder()))
		}
	}

	// 分页参数
	page := max(req.Page, 1)
	pageSize := min(max(req.PageSize, 10), 100) // 限定最大页大小
	offset := (page - 1) * pageSize

	// 查询总数
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	// 查询当前页
	users, err := q.Offset(int(offset)).Limit(int(pageSize)).All(ctx)
	if err != nil {
		return nil, err
	}

	// 构造响应
	var simpleUsers []*v1.SimpleUser
	for _, user := range users {
		simpleUsers = append(simpleUsers, convertSimpleUserToProto(user))
	}

	return &v1.ListUsersResponse{
		Result: true,
		Code:   200,
		Data: &v1.ListUsersResponse_PageResult{
			Total:    int32(total),
			Page:     page,
			PageSize: pageSize,
			Users:    simpleUsers,
		},
		Msg: "users retrieved successfully",
	}, nil
}
