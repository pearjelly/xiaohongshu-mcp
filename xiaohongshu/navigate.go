package xiaohongshu

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
)

type NavigateAction struct {
	page *rod.Page
}

func NewNavigate(page *rod.Page) *NavigateAction {
	return &NavigateAction{page: page}
}

func (n *NavigateAction) ToExplorePage(ctx context.Context) error {
	page := n.page.Context(ctx)

	page.MustNavigate("https://www.xiaohongshu.com/explore").
		MustWaitLoad().
		MustElement(`div#app`)

	return nil
}

func (n *NavigateAction) ToProfilePage(ctx context.Context) error {
	page := n.page.Context(ctx)

	// First navigate to explore page
	if err := n.ToExplorePage(ctx); err != nil {
		return err
	}

	// Find and click the "我" channel link in sidebar
	// Select specifically the 'user' component in sidebar
	profileLink, err := page.Timeout(5 * time.Second).Element(`div#app .side-bar .user`)
	if err != nil {
		// Fallback or detailed error
		return fmt.Errorf("could not find profile link in sidebar: %w", err)
	}
	profileLink.MustClick()

	// Wait for navigation to profile page (SPA navigation, so MustWaitLoad won't work)
	page.MustWait(`() => location.href.includes('/user/profile')`)

	return nil
}
