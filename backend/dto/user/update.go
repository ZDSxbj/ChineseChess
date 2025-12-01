package user

type UpdateUserRequest struct {
    Name   string  `json:"name"`
    Gender string  `json:"gender"`
    Email  string  `json:"email"`
    Avatar string  `json:"avatar"`
}

func (r *UpdateUserRequest) Examine() error {
    // 可添加字段校验逻辑
    return nil
}