package client

//
//import (
//	"context"
//	"fmt"
//	"git.internal.yunify.com/qxp/process/pkg/config"
//	"testing"
//)
//
//func TestIdentity_FindUsersByGroup(t *testing.T) {
//	var R []string
//	i, _ := NewIdentity(config.Config)
//	u, _ := i.FindUsersByGroup(context.Context(context.TODO()), "1")
//	//assert.Nil(t, err)
//	for _, us := range u.Users {
//		R = append(R, us.ID)
//	}
//	fmt.Println(R)
//}
//
//func TestIdentity_FindUserIDsByGroups(t *testing.T) {
//	i, _ := NewIdentity(config.Config)
//	g, _ := i.FindGroupsByUserID(context.Context(context.TODO()), "1ff0d06e-40e6-4256-aa9c-47577acfa9eb")
//	//assert.Nil(t, err)
//	fmt.Println(*g)
//}
//
//func init() {
//	config.Init("../../configs/config.yml")
//}
