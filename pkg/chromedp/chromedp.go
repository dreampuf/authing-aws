package chromedp

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"log"
	"net/url"
	"strconv"
)

type VisitAuthingOptions struct {
	URL, Username, Password, AppSelected string
	Debug, DisableHeadless               bool
	Logger                               *log.Logger
	ProfileDir                           string
}

type LoginPage struct {
	TabName, UsernameInput, PasswordInput, PasswordForm, PhoneCodeForm string
	LoginTabBtn, PasswordLoginBtn                                      string
}

type MainPage struct {
	LinkDiv string
}

type AuthingEntryPage struct {
	Version, URL, LoginPath string
	Login                   LoginPage
	Main                    MainPage
}

const AWS_SAML_ENDPOINT = "https://signin.aws.amazon.com/saml"
const AWS_CN_SAML_ENDPOINT = "https://signin.amazonaws.cn/saml"
const AWS_GOV_SAML_ENDPOINT = "https://signin.amazonaws-us-gov.com/saml"

func newEntryPage(url string) AuthingEntryPage {
	return AuthingEntryPage{
		Version:   "2.28.22",
		URL:       url,
		LoginPath: "/login",
		Login: LoginPage{
			TabName:          "Password",
			UsernameInput:    "input[type=text][placeholder*=\"username\" i]",
			PasswordInput:    "input[type=password][placeholder*=\"password\" i]",
			PasswordForm:     "#passworLogin",
			PasswordLoginBtn: "div.authing-ant-tabs-tabpane-active button.password",
			LoginTabBtn:      "#rc-tabs-0-tab-password",
			PhoneCodeForm:    "#phoneCode",
		},
		Main: MainPage{
			LinkDiv: "div[class^=styles_sortContainer]",
		},
	}
}

func VisitAuthing(ctx context.Context, opts VisitAuthingOptions) (string, error) {
	authing := newEntryPage(opts.URL)
	chromeOpts := append([]chromedp.ExecAllocatorOption{},
		chromedp.NoFirstRun,
		chromedp.DisableGPU,
		chromedp.NoDefaultBrowserCheck,
		chromedp.UserDataDir(opts.ProfileDir),
		chromedp.Flag("disable-extensions", true),
	)

	if !opts.DisableHeadless {
		chromeOpts = append(chromeOpts, chromedp.Headless)
	}

	logger := opts.Logger
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, chromeOpts...)
	defer cancel()

	// also set up a custom logger
	var (
		taskCtx    context.Context
		taskCancel context.CancelFunc
	)
	if opts.Debug {
		taskCtx, taskCancel = chromedp.NewContext(allocCtx, chromedp.WithDebugf(logger.Printf))
	} else {
		taskCtx, taskCancel = chromedp.NewContext(allocCtx)
	}
	defer taskCancel()
	var (
		samlresponse string
	)
	chromedp.ListenBrowser(taskCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *fetch.EventRequestPaused:
			go func() {
				c := chromedp.FromContext(taskCtx)
				if ev.Request.URL == AWS_CN_SAML_ENDPOINT ||
					ev.Request.URL == AWS_SAML_ENDPOINT ||
					ev.Request.URL == AWS_GOV_SAML_ENDPOINT {
					//logger.Printf("postdata: %s", ev.Request.PostData)
					if v, err := url.ParseQuery(ev.Request.PostData); err == nil {
						samlresponse = v.Get("SAMLResponse")
					}
				}
				if err := fetch.ContinueRequest(ev.RequestID).Do(cdp.WithExecutor(taskCtx, c.Browser)); err != nil {
					logger.Printf("Fetch Failed due to: %s", err)
				}
			}()
		}
	})

	// tab btn: div[class^="styles_authing-tab-item"],div[class*=" styles_authing-tab-item"]:last-child
	// ensure that the browser process is started
	var appItemNodes []*cdp.Node
	appItemsMap := map[string]*cdp.Node{}
	appItemsName := []string{}
	if err := chromedp.Run(
		taskCtx,
		chromedp.Navigate(authing.URL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			c := chromedp.FromContext(ctx)
			// note that the executor is "Browser" so that it will emit events
			// for all targets.
			return fetch.Enable().Do(cdp.WithExecutor(ctx, c.Browser))
		}),
		chromedp.WaitNotPresent("div[class^=styles_g2-init-setting-loading]", chromedp.ByQuery),
		chromedp.WaitVisible(fmt.Sprintf("%s,%s,%s", authing.Login.PasswordForm, authing.Login.PhoneCodeForm, authing.Main.LinkDiv), chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var nodes []*cdp.Node
			if err := chromedp.Nodes(authing.Login.PasswordForm, &nodes, chromedp.AtLeast(0), chromedp.ByQuery).Do(ctx); err != nil {
				return err
			}
			if len(nodes) == 0 {
				return nil
			}
			//if err := chromedp.Click(authing.Login.LoginTabBtn, chromedp.ByID).Do(ctx); err != nil {
			//	return err
			//}
			if err := chromedp.SendKeys(authing.Login.UsernameInput, opts.Username, chromedp.ByQuery).Do(ctx); err != nil {
				return err
			}
			if err := chromedp.SendKeys(authing.Login.PasswordInput, opts.Password, chromedp.ByQuery).Do(ctx); err != nil {
				return err
			}
			return chromedp.Click(authing.Login.PasswordLoginBtn, chromedp.ByQuery).Do(ctx)
		}),
		chromedp.WaitVisible(authing.Main.LinkDiv),
		chromedp.Nodes("div[class^=styles_appItem]", &appItemNodes, chromedp.ByQueryAll),
	); err != nil {
		return "", fmt.Errorf("login into failed: %w", err)
	}

	// selecting app
	for _, i := range appItemNodes {
		var appItemHTML string
		var appImgs []*cdp.Node
		if err := chromedp.Run(taskCtx, chromedp.Nodes("img[src*=aws]", &appImgs, chromedp.AtLeast(0), chromedp.ByQuery, chromedp.FromNode(i))); err != nil {
			return "", fmt.Errorf("logging for app items failed: %w", err)
		}
		if len(appImgs) == 0 {
			continue
		}
		if err := chromedp.Run(taskCtx, chromedp.InnerHTML("p[class^=styles_appName]", &appItemHTML, chromedp.ByQuery, chromedp.FromNode(i))); err != nil {
			return "", fmt.Errorf("logging for app items failed: %w", err)
		}
		appItemsMap[appItemHTML] = i
		appItemsName = append(appItemsName, appItemHTML)
	}

	var (
		item       *cdp.Node
		ok         bool
		n_selected int
		nerr       error
	)
	n_selected, nerr = strconv.Atoi(opts.AppSelected)
	if nerr == nil {
		if n_selected >= len(appItemsName) {
			return "", fmt.Errorf("can't be more than %d. n: %d", len(appItemsName), n_selected)
		}
		item, ok = appItemsMap[appItemsName[n_selected]]
	} else {
		item, ok = appItemsMap[opts.AppSelected]
	}
	if !ok {
		logger.Printf("don't find select app \"%s\":", opts.AppSelected)
		for n, appItemName := range appItemsName {
			logger.Printf("%02d. \"%s\"", n, appItemName)
		}
		return "", fmt.Errorf("app selection doesn't match: %s", opts.AppSelected)
	}

	newTabCh := chromedp.WaitNewTarget(taskCtx, func(info *target.Info) bool {
		return info.URL != ""
	})
	if err := chromedp.Run(
		taskCtx,
		chromedp.Click([]cdp.NodeID{item.NodeID}, chromedp.ByNodeID),
	); err != nil {
		return "", fmt.Errorf("click to app failed: %w", err)
	}

	//newTabCtx, newTabCancel := context.WithCancel(taskCtx)
	newTabCtx, newTabCancel := chromedp.NewContext(taskCtx,
		chromedp.WithTargetID(<-newTabCh),
	)
	defer newTabCancel()

	if err := chromedp.Run(
		newTabCtx,
		chromedp.WaitVisible("#nav-logo", chromedp.ByID),
	); err != nil {
		return "", fmt.Errorf("#nav-logo isn't exist: %w", err)
	}
	return samlresponse, nil
}
