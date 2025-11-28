package _123Share

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"

	_123 "github.com/OpenListTeam/OpenList/v4/drivers/123"
	"github.com/OpenListTeam/OpenList/v4/drivers/base"
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type Pan123Share struct {
	model.Storage
	Addition
	apiRateLimit sync.Map
	ref          *_123.Pan123
}

func (d *Pan123Share) Config() driver.Config {
	return config
}

func (d *Pan123Share) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Pan123Share) Init(ctx context.Context) error {
	return nil
}

func (d *Pan123Share) InitReference(storage driver.Driver) error {
	refStorage, ok := storage.(*_123.Pan123)
	if ok {
		d.ref = refStorage
		return nil
	}
	return fmt.Errorf("ref: storage is not 123Pan")
}

func (d *Pan123Share) Drop(ctx context.Context) error {
	d.ref = nil
	return nil
}

func (d *Pan123Share) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(ctx, dir.GetID())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return src, nil
	})
}

// 关键修改：支持分片下载绕过 1GB 限制
func (d *Pan123Share) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	if f, ok := file.(File); ok {
		data := base.Json{
			"shareKey":  d.ShareKey,
			"SharePwd":  d.SharePwd,
			"etag":      f.Etag,
			"fileId":    f.FileId,
			"s3keyFlag": f.S3KeyFlag,
			"size":      f.Size,
		}

		resp, err := d.request(DownloadInfo, http.MethodPost, func(req *resty.Request) {
			req.SetBody(data)
		}, nil)
		if err != nil {
			return nil, err
		}

		downloadUrl := utils.Json.Get(resp, "data", "DownloadURL").ToString()
		if downloadUrl == "" {
			return nil, fmt.Errorf("DownloadURL is empty")
		}

		ou, err := url.Parse(downloadUrl)
		if err != nil {
			return nil, err
		}

		// 处理 base64 加密的 params 参数
		u_ := ou.String()
		nu := ou.Query().Get("params")
		if nu != "" {
			du, err := base64.StdEncoding.DecodeString(nu)
			if err == nil {
				if u, err := url.Parse(string(du)); err == nil {
					u_ = u.String()
				}
			}
		}

		log.Debug("final download url: ", u_)

		// 关键：构造支持 Range 的 Link
		link := &model.Link{
			URL: u_,
			Header: http.Header{
				"Referer": []string{"https://www.123pan.com/"},
			},
		}

		// 强制声明支持分片下载（OpenList 下载器看到这个就会自动多线程）
		link.Header.Set("Accept-Ranges", "bytes")

		// 强烈建议加上 Content-Length，让进度条准确
		if f.Size > 0 {
			link.Header.Set("Content-Length", strconv.FormatInt(f.Size, 10))
		}

		// 可选：加上 Content-Type（部分播放器需要）
		if mime := utils.GetMimeType(f.GetName()); mime != "" {
			link.Header.Set("Content-Type", mime)
		}

		return link, nil
	}
	return nil, fmt.Errorf("can't convert obj")
}

func (d *Pan123Share) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return errs.NotSupport
}

func (d *Pan123Share) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotSupport
}

func (d *Pan123Share) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return errs.NotSupport
}

func (d *Pan123Share) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotSupport
}

func (d *Pan123Share) Remove(ctx context.Context, obj model.Obj) error {
	return errs.NotSupport
}

func (d *Pan123Share) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	return errs.NotSupport
}

func (d *Pan123Share) APIRateLimit(ctx context.Context, api string) error {
	value, _ := d.apiRateLimit.LoadOrStore(api,
		rate.NewLimiter(rate.Every(700*time.Millisecond), 1))
	limiter := value.(*rate.Limiter)
	return limiter.Wait(ctx)
}

var _ driver.Driver = (*Pan123Share)(nil)
