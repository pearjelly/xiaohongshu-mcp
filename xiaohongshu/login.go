package xiaohongshu

import (
	"context"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
)

type LoginAction struct {
	page *rod.Page
}

func NewLogin(page *rod.Page) *LoginAction {
	return &LoginAction{page: page}
}

func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error) {
	pp := a.page.Context(ctx)

	// Check if we are already on xiaohongshu.com
	info, err := pp.Info()
	if err == nil && (info.URL == "" || info.URL == "about:blank") {
		// Only navigate if we are on a blank page
		if err := pp.Navigate("https://www.xiaohongshu.com/explore"); err != nil {
			return false, errors.Wrap(err, "navigate to explore page failed")
		}
		if err := pp.WaitLoad(); err != nil {
			return false, errors.Wrap(err, "wait page load failed")
		}
		time.Sleep(1 * time.Second)
	}

	// Try to find the element on the current page without refreshing
	// Using a more generic selector for the user profile entry in sidebar which exists on most pages
	exists, _, err := pp.Has(`.side-bar .user`)
	if err != nil {
		// Fallback to original selector if checking fails (though Has shouldn't fail easily)
		exists, _, err = pp.Has(`.main-container .user .link-wrapper .channel`)
		if err != nil {
			return false, errors.Wrap(err, "check login status failed")
		}
	}

	if exists {
		return true, nil
	}

	// Double check: if we are on XHS but element not found, maybe we need to wait or it's a different page layout?
	// But purely for "CheckStatus", returning false is safer than forcibly refreshing which disrupts user.
	// If the user is logged out, the element won't be there.

	// One edge case: We are on XHS but the DOM hasn't loaded the sidebar yet.
	// But this is a background check.

	return false, nil
}

func (a *LoginAction) Login(ctx context.Context) error {
	pp := a.page.Context(ctx)

	// 导航到小红书首页，这会触发二维码弹窗
	if err := pp.Navigate("https://www.xiaohongshu.com/explore"); err != nil {
		return errors.Wrap(err, "navigate to explore page failed")
	}

	if err := pp.WaitLoad(); err != nil {
		return errors.Wrap(err, "wait page load failed")
	}

	// 等待一小段时间让页面完全加载
	time.Sleep(2 * time.Second)

	// 检查是否已经登录
	if exists, _, _ := pp.Has(".main-container .user .link-wrapper .channel"); exists {
		// 已经登录，直接返回
		return nil
	}

	// 等待扫码成功提示或者登录完成
	// 这里我们等待登录成功的元素出现，这样更简单可靠
	_, err := pp.Element(".main-container .user .link-wrapper .channel")
	if err != nil {
		return errors.Wrap(err, "wait for login element failed")
	}

	return nil
}

func (a *LoginAction) FetchQrcodeImage(ctx context.Context) (string, bool, error) {
	pp := a.page.Context(ctx)

	// 导航到小红书首页，这会触发二维码弹窗
	if err := pp.Navigate("https://www.xiaohongshu.com/explore"); err != nil {
		return "", false, errors.Wrap(err, "navigate to explore page failed")
	}

	if err := pp.WaitLoad(); err != nil {
		return "", false, errors.Wrap(err, "wait page load failed")
	}

	// 等待一小段时间让页面完全加载
	time.Sleep(2 * time.Second)

	// 检查是否已经登录
	if exists, _, _ := pp.Has(".main-container .user .link-wrapper .channel"); exists {
		return "", true, nil
	}

	// 获取二维码图片
	el, err := pp.Element(".login-container .qrcode-img")
	if err != nil {
		return "", false, errors.Wrap(err, "find qrcode element failed")
	}

	src, err := el.Attribute("src")
	if err != nil {
		return "", false, errors.Wrap(err, "get qrcode src failed")
	}
	if src == nil || len(*src) == 0 {
		return "", false, errors.New("qrcode src is empty")
	}

	return *src, false, nil
}

func (a *LoginAction) WaitForLogin(ctx context.Context) bool {
	pp := a.page.Context(ctx)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			el, err := pp.Element(".main-container .user .link-wrapper .channel")
			if err == nil && el != nil {
				return true
			}
		}
	}
}
