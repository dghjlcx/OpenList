// drivers/189share/driver.go
package _189share

import (
    "context"
    "regexp"
    "github.com/OpenListTeam/OpenList/v4/internal/driver"
    "github.com/OpenListTeam/OpenList/v4/internal/model"
    "github.com/OpenListTeam/OpenList/v4/internal/errs"
    "github.com/OpenListTeam/OpenList/v4/drivers/_189pc" // 引用原驱动包
)

type Share189 struct {
    model.Storage
    Addition

    master *_189pc.Cloud189PC // 主驱动引用
    shareID string
}

type Addition struct {
    ShareLink     string `json:"share_link" required:"true"`
    SharePassword string `json:"share_password"`
    Reference     string `json:"reference" help:"选择你的 189CloudPC 账号作为认证来源"`
}

var config = driver.Config{
    Name: "189Share",
    OnlyLocal: false,
    NoUpload: true,
    NoMkdir: true,
}

func (s *Share189) Config() driver.Config { return config }

func (s *Share189) Init(ctx context.Context) error {
    // 解析 shareID
    re := regexp.MustCompile(`/t/([a-zA-Z0-9]+)`)
    if m := re.FindStringSubmatch(s.Addition.ShareLink); len(m) > 1 {
        s.shareID = m[1]
    } else {
        return errs.NewErr("分享链接格式错误")
    }

    // 引用主驱动（关键！）
    if err := op.GetStorageByName(s.Addition.Reference, &s.master); err != nil {
        return errs.NewErr("未找到引用的 189CloudPC 账号，请先添加")
    }
    if s.master.GetTokenInfo() == nil {
        return errs.NewErr("主账号未登录")
    }
    return nil
}

// List：复用主驱动的 request 方法，但换分享接口
func (s *Share189) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
    return s.master.ListShareFiles(ctx, s.shareID, dir.GetID(), s.Addition.SharePassword)
}

// Link：获取分享直链
func (s *Share189) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
    return s.master.GetShareDownloadUrl(ctx, s.shareID, file.GetID())
}

// 禁用所有写操作
func (s *Share189) MakeDir(...) error { return errs.NotSupport }
func (s *Share189) Put(...) error { return errs.NotSupport }
func (s *Share189) Remove(...) error { return errs.NotSupport }
// ... 其他同理

func init() {
    op.RegisterDriver(func() driver.Driver { return &Share189{} })
}