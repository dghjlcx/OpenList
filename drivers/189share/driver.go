// drivers/189share/driver.go
package _189share

import (
	"context"
	"regexp"
	"strings"

	"github.com/OpenListTeam/OpenList/v4/drivers/189pc" // 正确 import 官方包名
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
)

type Share189 struct {
	model.Storage
	Addition

	master  *Cloud189PC // 注意这里是 *Cloud189PC，不是 *_189pc.Cloud189PC
	shareID string
}

type Addition struct {
	ShareLink     string `json:"share_link" required:"true" help:"天翼云盘分享链接，如 https://cloud.189.cn/t/xxxxxx"`
	SharePassword string `json:"share_password" help:"分享提取码（若有）"`
	Reference     string `json:"reference" required:"true" type:"select" options:"storages" help:"选择已登录的 189CloudPC 账号"`
}

var config = driver.Config{
	Name:        "189Share",
	DefaultRoot: "/",
	NoUpload:    true,
	NoMkdir:     true,
	OnlyLocal:   false,
	CheckStatus: true,
	Alert:       "只读模式 · 使用你的 189 账号无限速访问他人分享",
}

func (s *Share189) Config() driver.Config { return config }
func (s *Share189) GetAddition() driver.Additional { return &s.Addition }

func (s *Share189) Init(ctx context.Context) error {
	re := regexp.MustCompile(`/t/([a-zA-Z0-9]+)`)
	if m := re.FindStringSubmatch(s.Addition.ShareLink); len(m) < 2 {
		return errs.NewErr("分享链接格式错误")
	}
	s.shareID = m[1]

	if err := op.GetStorageByName(s.Addition.Reference, &s.master); err != nil || s.master == nil {
		return errs.NewErr("找不到引用的 189CloudPC 账号")
	}
	if s.master.GetTokenInfo() == nil {
		return errs.NewErr("主账号未登录")
	}
	return nil
}

func (s *Share189) Drop(ctx context.Context) error { return nil }

func (s *Share189) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	folderID := "0"
	if dir != nil && dir.GetID() != "" && dir.GetID() != "/" {
		folderID = dir.GetID()
	}
	return s.master.ListShareFiles(ctx, s.shareID, folderID, s.Addition.SharePassword)
}

func (s *Share189) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	return s.master.GetShareDownloadUrl(ctx, s.shareID, file.GetID())
}

// 禁用写操作
func (s *Share189) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error)  { return nil, errs.NotSupport }
func (s *Share189) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error)  { return nil, errs.NotSupport }
func (s *Share189) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) { return nil, errs.NotSupport }
func (s *Share189) Copy(ctx context.Context, srcObj, dstDir model.Obj) error          { return errs.NotSupport }
func (s *Share189) Remove(ctx context.Context, obj model.Obj) error                 { return errs.NotSupport }
func (s *Share189) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	return nil, errs.NotSupport
}

func (s *Share189) GetStorageName() string {
	name := "189分享"
	if s.MountPath != "" {
		name += " - " + strings.Trim(s.MountPath, "/")
	}
	if s.Addition.SharePassword != "" {
		name += " [密码]"
	}
	return name
}

func init() {
	op.RegisterDriver(func() driver.Driver { return &Share189{} })
}
