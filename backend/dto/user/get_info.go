package user

import (
	"fmt"
)

type GetUserInfoRequest struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (r *GetUserInfoRequest) Examine() error {
	if r.Id <= 0 && r.Name == "" {
		return fmt.Errorf("用户ID或用户名不能为空")
	}
	return nil
}

type GetUserInfoResponse struct {
	UserInfo
}
