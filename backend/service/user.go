package service

import (
	"chinese-chess-backend/database"
	dto "chinese-chess-backend/dto/user"
	recordModel "chinese-chess-backend/model/record"
	userModel "chinese-chess-backend/model/user"
	"chinese-chess-backend/utils"
	"os"
	"time"

	"gorm.io/gorm" // 新增gorm包导入

	"errors"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

// AddUserExp 为指定用户增加经验值（可为负，但最小不低于0）
func (us *UserService) AddUserExp(userID int, delta int) error {
	if delta == 0 {
		return nil
	}
	db := database.GetMysqlDb()
	var u userModel.User
	if err := db.Where("id = ?", userID).First(&u).Error; err != nil {
		return err
	}
	newExp := u.Exp + delta
	if newExp < 0 {
		newExp = 0
	}
	u.Exp = newExp
	return db.Save(&u).Error
}

// UpdateUserStats 重新计算并更新用户的总场次与胜率
// 统计规则：
// - 本地对战不入库，天然不计入
// - 人机对战（game_type=1）：仅当分出胜负（result=0或1）才计入场次；和棋不计入场次
// - 随机匹配（game_type=0）：一局结束即计入场次（含和棋）
// - 胜率基于胜/负计算：wins / (wins + losses) * 100，和棋不影响分母
func (us *UserService) UpdateUserStats(userID int) error {
	db := database.GetMysqlDb()

	var wins int64
	var losses int64
	var randomFriendAll int64
	var randomFriendDraws int64
	var aiWinLoss int64

	// 胜场：用户为红且result=0，或用户为黑且result=1；统计随机匹配、好友、人机
	if err := db.Model(&recordModel.GameRecord{}).
		Where("game_type IN ?", []int{0, 1, 2}).
		Where("(red_id = ? AND result = 0) OR (black_id = ? AND result = 1)", userID, userID).
		Count(&wins).Error; err != nil {
		return err
	}

	// 负场：用户为红且result=1，或用户为黑且result=0；统计随机匹配、好友、人机
	if err := db.Model(&recordModel.GameRecord{}).
		Where("game_type IN ?", []int{0, 1, 2}).
		Where("(red_id = ? AND result = 1) OR (black_id = ? AND result = 0)", userID, userID).
		Count(&losses).Error; err != nil {
		return err
	}

	// 随机匹配 + 好友对战总局数：包含和棋
	if err := db.Model(&recordModel.GameRecord{}).
		Where("game_type IN ?", []int{0, 2}).
		Where("red_id = ? OR black_id = ?", userID, userID).
		Count(&randomFriendAll).Error; err != nil {
		return err
	}

	// 随机匹配 + 好友对战的和棋数
	if err := db.Model(&recordModel.GameRecord{}).
		Where("game_type IN ?", []int{0, 2}).
		Where("red_id = ? OR black_id = ?", userID, userID).
		Where("result = 2").
		Count(&randomFriendDraws).Error; err != nil {
		return err
	}

	// 人机计入的局数：仅胜/负（result != 2）
	if err := db.Model(&recordModel.GameRecord{}).
		Where("game_type = 1").
		Where("(red_id = ? OR black_id = ?)", userID, userID).
		Where("result IN (0,1)").
		Count(&aiWinLoss).Error; err != nil {
		return err
	}

	// 总场次：随机匹配+好友 全部；人机仅胜负
	totalGames := int(randomFriendAll + aiWinLoss)
	// 胜率分母：随机匹配+好友（胜+负+和），人机（胜+负）——与总场次等价
	denom := float64(wins + losses + randomFriendDraws)
	winRate := 0.0
	if denom > 0 {
		winRate = float64(wins) * 100.0 / denom
	}

	// 更新用户数据
	var u userModel.User
	if err := db.Where("id = ?", userID).First(&u).Error; err != nil {
		return err
	}
	u.TotalGames = totalGames
	u.WinRate = winRate
	return db.Save(&u).Error
}

func (us *UserService) Register(req *dto.RegisterRequest) (dto.RegisterResponse, error) {
	var registerResp dto.RegisterResponse
	var err error

	// 统一头像 URL 前缀，支持环境变量 PUBLIC_API_PREFIX，默认 http://localhost:8080/api
	apiBase := os.Getenv("PUBLIC_API_PREFIX")
	if apiBase == "" {
		apiBase = "http://localhost:8080/api"
	}
	defaultAvatar := apiBase + "/uploads/avatars/default.png"

	user := userModel.User{
		Name:  req.Name,
		Email: req.Email,
		// 默认存储完整 URL，避免前端环境差异
		Avatar: defaultAvatar,
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
		var count int64
		db.Model(&userModel.User{}).
			Where("name = ? AND id <> ?", req.Name, userID).
			Count(&count)
		if count > 0 {
			return errors.New("该id已被注册")
		}
		user.Name = req.Name
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Avatar != "" {
		// 兼容：当前端上传返回相对路径时，统一补齐为完整 URL
		apiBase := os.Getenv("PUBLIC_API_PREFIX")
		if apiBase == "" {
			apiBase = "http://localhost:8080/api"
		}
		if len(req.Avatar) > 0 && req.Avatar[0] == '/' { // 以相对路径开头
			user.Avatar = apiBase + req.Avatar
		} else {
			user.Avatar = req.Avatar
		}
	}

	return db.Save(&user).Error
}

func (us *UserService) UpdateEmailWithCode(userID int, email string, code string) error {
	// 新增：检查邮箱是否已被其他用户使用
	db := database.GetMysqlDb()
	var existingUser userModel.User
	if err := db.Where("email = ? AND id != ?", email, userID).First(&existingUser).Error; err == nil {
		return errors.New("该邮箱已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 处理数据库查询错误
		return errors.New("检查邮箱失败")
	}
	// 校验验证码
	realCode, err := database.GetValue(email)
	if err != nil {
		return errors.New("验证码错误或已过期")
	}
	if realCode != code {
		return errors.New("验证码错误")
	}

	// 修改邮箱
	var user userModel.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.New("用户不存在")
	}
	user.Email = email
	return db.Save(&user).Error
}

func (us *UserService) UpdatePassword(userID int, oldPassword string, newPassword string) error {
	db := database.GetMysqlDb()
	var user userModel.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.New("用户不存在")
	}
	// 校验旧密码
	if !utils.CheckPassword(user.Password, oldPassword) {
		return errors.New("密码错误")
	}
	// 加密新密码并保存
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.Password = hashed
	return db.Save(&user).Error
}

// CheckPassword 仅校验原密码是否正确
func (us *UserService) CheckPassword(userID int, password string) error {
	db := database.GetMysqlDb()
	var user userModel.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.New("用户不存在")
	}
	if !utils.CheckPassword(user.Password, password) {
		return errors.New("密码错误")
	}
	return nil
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
		// 人机对战时，对手 ID 为 0，显示 AI 难度
		if record.GameType == 1 && opponentID == 0 {
			var levelLabel string
			switch {
			case record.AILevel <= 2:
				levelLabel = "简单"
			case record.AILevel <= 4:
				levelLabel = "中等"
			default:
				levelLabel = "困难"
			}
			opponentName = "AI (" + levelLabel + ")"
		} else if opponentName == "" {
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
			AILevel:      record.AILevel,
		}

		response.Records = append(response.Records, item)
	}

	return &response, nil
}
