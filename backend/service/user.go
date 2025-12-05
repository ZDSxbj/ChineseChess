package service

import (
	"chinese-chess-backend/database"
	dto "chinese-chess-backend/dto/user"
	recordModel "chinese-chess-backend/model/record"
	userModel "chinese-chess-backend/model/user"
	"chinese-chess-backend/utils"
	"time"

	"errors"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (us *UserService) Register(req *dto.RegisterRequest) (dto.RegisterResponse, error) {
	var registerResp dto.RegisterResponse
	var err error

	user := userModel.User{
		Name:  req.Name,
		Email: req.Email,
	}

	// 校验邮箱是否已注册
	db := database.GetMysqlDb()
	if err = db.Where("email = ?", req.Email).First(&user).Error; err == nil {
		return registerResp, errors.New("邮箱已注册")
	}

	code, err := database.GetValue(req.Email)

	if err != nil {
		return registerResp, err
	}

	if code != req.VCode {
		err = errors.New("验证码错误")
		return registerResp, err
	}

	hashPass, err := utils.HashPassword(req.Password)
	if err != nil {
		return registerResp, err
	}

	user.Password = hashPass

	db.Create(&user)

	// 签发token
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return registerResp, err
	}

	registerResp.Token = token
	registerResp.Name = user.Name

	return registerResp, nil
}

func (us *UserService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	var loginResp dto.LoginResponse
	var err error

	// 查询用户
	db := database.GetMysqlDb()
	user := userModel.User{}
	if err = db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 校验密码
	if !utils.CheckPassword(user.Password, req.Password) {
		return nil, errors.New("密码错误")
	}

	// 签发token
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	loginResp.Token = token
	loginResp.Name = user.Name

	// 更新在线状态为 true
	if err := db.Model(&user).Update("online", true).Error; err != nil {
		// 不阻塞登录流程，但记录错误
		// 可以改为返回错误以强制要求成功写入在线状态
	}

	return &loginResp, nil
}

func (us *UserService) SendVCode(req *dto.SendVCodeRequest) error {
	var err error

	// 生成验证码
	code := utils.GetRandomCode()

	// 发送验证码
	err = utils.SendMail(req.Email, "验证码", code)
	if err != nil {
		return err
	}

	// 存储验证码到数据库
	err = database.SetValue(req.Email, code, time.Minute*5)
	if err != nil {
		return err
	}

	return nil
}

func (uc *UserService) GetUserInfo(req *dto.GetUserInfoRequest) (*dto.GetUserInfoResponse, error) {
	var userInfoResp dto.GetUserInfoResponse
	var err error

	// 查询用户
	db := database.GetMysqlDb()
	user := userModel.User{}

	query := db
	if req.Id > 0 {
		query = query.Where("id = ?", req.Id)
	} else if req.Name != "" {
		query = query.Where("name = ?", req.Name)
	}

	if err = query.First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	userInfoResp.UserInfo = dto.UserInfo{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		Exp:        user.Exp,
		Avatar:     user.Avatar,
		Gender:     user.Gender,
		TotalGames: user.TotalGames,
		WinRate:    user.WinRate,
	}

	return &userInfoResp, nil
}

func (us *UserService) UpdateUserInfo(userID int, req *dto.UpdateUserRequest) error {
	db := database.GetMysqlDb()
	var user userModel.User

	// 查询用户是否存在
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 更新字段（只更新非空字段）
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	return db.Save(&user).Error
}

func (us *UserService) GetGameRecords(req *dto.GetGameRecordsRequest) (*dto.GetGameRecordsResponse, error) {
	db := database.GetMysqlDb()
	var records []recordModel.GameRecord
	var response dto.GetGameRecordsResponse

	// 查询该用户的所有对局记录（作为红方或黑方）
	if err := db.Where("red_id = ? OR black_id = ?", req.UserID, req.UserID).
		Order("start_time DESC").
		Find(&records).Error; err != nil {
		return nil, errors.New("查询对局记录失败")
	}

	// 收集所有对手的 ID
	var opponentIDs []uint
	opponentIDMap := make(map[uint]int) // 记录对手ID在records中的位置

	for i, record := range records {
		var opponentID uint
		if record.RedID == uint(req.UserID) {
			opponentID = record.BlackID
		} else {
			opponentID = record.RedID
		}
		opponentIDs = append(opponentIDs, opponentID)
		opponentIDMap[opponentID] = i
	}

	// 批量查询对手信息
	var opponents []userModel.User
	opponentMap := make(map[uint]string)

	if len(opponentIDs) > 0 {
		if err := db.Model(&userModel.User{}).
			Select("id, name").
			Where("id IN ?", opponentIDs).
			Find(&opponents).Error; err != nil {
			return nil, errors.New("查询对手信息失败")
		}

		for _, opponent := range opponents {
			opponentMap[opponent.ID] = opponent.Name
		}
	}

	// 构建返回结果
	response.Records = make([]dto.GameRecordItem, 0, len(records))

	for _, record := range records {
		isRed := record.RedID == uint(req.UserID)
		var opponentID uint
		var result int

		if isRed {
			opponentID = record.BlackID
			// 当前用户是红方
			if record.Result == 0 {
				result = 0 // 红方胜
			} else if record.Result == 1 {
				result = 1 // 红方负（黑方胜）
			} else {
				result = 2 // 和棋
			}
		} else {
			opponentID = record.RedID
			// 当前用户是黑方
			if record.Result == 1 {
				result = 0 // 黑方胜
			} else if record.Result == 0 {
				result = 1 // 黑方负（红方胜）
			} else {
				result = 2 // 和棋
			}
		}

		// 计算总步数（兼容多种后端存储格式）
		// 1) 如果 history 是 JSON 格式：可能是 moves 数组（每项包含 from/to）或 positions 数组（每项为 {x,y}）
		// 2) 如果 history 是紧凑数字字符串（如 "6665" 表示 (6,6)->(6,5)），则按字符解析
		totalSteps := 0

		hs := record.History
		l := len(hs)
		if l >= 4 {
			totalSteps = (l / 2) / 2
		}

		opponentName := opponentMap[opponentID]
		if opponentName == "" {
			opponentName = "未知玩家"
		}

		item := dto.GameRecordItem{
			ID:           record.ID,
			OpponentID:   opponentID,
			OpponentName: opponentName,
			Result:       result,
			GameType:     record.GameType,
			IsRed:        isRed,
			TotalSteps:   totalSteps,
			History:      record.History,
			StartTime:    record.StartTime,
		}

		response.Records = append(response.Records, item)
	}

	return &response, nil
}
