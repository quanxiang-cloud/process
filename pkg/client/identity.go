package client

import (
	"context"
	"github.com/quanxiang-cloud/process/internal"
	"github.com/quanxiang-cloud/process/pkg/config"
	"github.com/quanxiang-cloud/process/pkg/misc/client"
	"net/http"
	"strings"
)

const (
	// OrgHost host
	OrgHost = "http://org"
)

// Identity interface
type Identity interface {
	FindUsersByGroup(ctx context.Context, groupID string) (*DepUsersResp, error)
	FindUsersByGroups(ctx context.Context, groupIDs []string) (*DepUsersResp, error)
	FindUserIDsByGroups(ctx context.Context, groupIDs []string) ([]string, error)
	FindGroupsByUserID(ctx context.Context, userID string) (*Group, error)
}

type identity struct {
	getGroupsByUser  string
	getUsersByGroups string
	client           http.Client
}

// NewIdentity init
func NewIdentity(conf *config.Configs) (Identity, error) {
	i := &identity{
		client: client.New(conf.InternalNet),
		//getGroupsByUser:  OrgHost + "/api/v1/org/usersInfo",
		//getUsersByGroups: OrgHost + "/api/v1/org/otherGetUserList",
		getGroupsByUser:  OrgHost + "/api/v1/org/o/user/ids",
		getUsersByGroups: OrgHost + "/api/v1/org/o/user/dep/id",
	}
	return i, nil
}

// FindUsersByGroup find users by group
func (i *identity) FindUsersByGroup(ctx context.Context, groupID string) (*DepUsersResp, error) {
	var resp *DepUsersResp
	params := map[string]interface{}{"depID": groupID, "includeChildDEPChild": 0}
	err := client.POST(ctx, &i.client, i.getUsersByGroups, params, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// FindUsersByGroups find users by groups
func (i *identity) FindUsersByGroups(ctx context.Context, groupIDs []string) (*DepUsersResp, error) {
	var resp *DepUsersResp
	err := client.POST(ctx, &i.client, i.getUsersByGroups, groupIDs, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// FindUserIDsByGroups find userIDs by groups
func (i *identity) FindUserIDsByGroups(ctx context.Context, groupIDs []string) ([]string, error) {
	users, err := i.FindUsersByGroups(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	var userIDs []string
	if len(users.Users) > 0 {
		for _, value := range users.Users {
			userIDs = append(userIDs, value.ID)
		}
	}
	return userIDs, nil
}

func (i *identity) FindGroupsByUserID(ctx context.Context, userID string) (*Group, error) {
	var resp *UsersInfosResp
	userIDs := map[string][]string{"ids": {userID}}
	err := client.POST(ctx, &i.client, i.getGroupsByUser, userIDs, &resp)
	if err != nil {
		return nil, err
	}
	if len(resp.Users) > 0 && len(resp.Users[0].Deps) > 0 && len(resp.Users[0].Deps[0]) > 0 {
		dep := resp.Users[0].Deps[0][0]
		group := &Group{
			ID:   strings.Join([]string{internal.Dep, "_", dep.ID}, ""),
			Name: dep.Name,
			Type: internal.Dep,
		}
		return group, nil
	}
	return nil, nil
}

// User info
type User struct {
	ID        string `json:"id"`
	UserName  string `json:"userName"`
	UseStatus int    `json:"useStatus"`
	TenantID  string `json:"tenantId"`
}

// Group info
type Group struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	TenantID string `json:"tenantId"`
}

// UsersInfosResp 用户信息列表
type UsersInfosResp struct {
	Users []*UserInfo `json:"users"`
}

// UserInfo 单个用户信息
type UserInfo struct {
	CommonUserInfo
	Deps    [][]*Dep            `json:"deps"`
	Leaders [][]*DepUserLeaders `json:"leaders"`
	Status  int                 `json:"status"`
}

// DepUsersResp 部门下属人员列表
type DepUsersResp struct {
	Users []*DepUserInfo `json:"users"`
}

// DepUserInfo 部门下属人员信息
type DepUserInfo struct {
	CommonUserInfo
	Leaders [][]*DepUserLeaders `json:"leaders"`
	Status  int                 `json:"status"`
}

// DepUserLeaders 领导信息
type DepUserLeaders struct {
	CommonUserInfo
}

// CommonUserInfo 单个用户能用信息
type CommonUserInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	UseStatus int    `json:"useStatus"` // 1:正常，-2禁用,-1真删除
	Position  string `json:"position"`
}

// Dep 部门信息
type Dep struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Pid     string `json:"pid"`
	SuperID string `json:"superID"`
	Grade   int    `json:"grade"`
}
