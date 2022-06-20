package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dreampuf/authing-aws/pkg/aws"
	"github.com/dreampuf/authing-aws/pkg/chromedp"
	"github.com/dreampuf/authing-aws/pkg/saml"
	"log"
	"os"
	"path"
)

var (
	flag_username        = flag.String("username", "", "username")
	flag_password        = flag.String("password", "", "password")
	flag_url             = flag.String("url", "", "URL")
	flag_app_selected    = flag.String("app", "", "selected app")
	flag_duration        = flag.Int64("duration", 10*60*60, "duration in seconds")
	flag_region          = flag.String("region", "cn-north-1", "region of SAMLResponse")
	flag_debug           = flag.Bool("debug", false, "enable debug logs")
	flag_disableheadless = flag.Bool("disable-headless", false, "disable headless mode to show chrome")

	// goreleaser
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func fetchCreds(logger *log.Logger) error {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("can't locate home dir of user: %w", err)
	}

	authing_dir_name := ".authing"
	authing_dir := path.Join(dirname, authing_dir_name)
	if _, err := os.Stat(authing_dir); os.IsNotExist(err) {
		if err = os.MkdirAll(authing_dir, 0700); err != nil {
			return fmt.Errorf("creating ~/%s failed: %w", authing_dir_name, err)
		}
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	samlresponse, err := chromedp.VisitAuthing(ctx, chromedp.VisitAuthingOptions{
		URL:             *flag_url,
		Username:        *flag_username,
		Password:        *flag_password,
		AppSelected:     *flag_app_selected,
		Debug:           *flag_debug,
		DisableHeadless: *flag_disableheadless,
		Logger:          logger,
		ProfileDir:      authing_dir,
	})
	if err != nil {
		return err
	}
	if samlresponse == "" {
		return fmt.Errorf("invalidated samlresponse")
	}
	samlassertion, err := saml.DecodeBase64edSAMLResponse(samlresponse)
	if err != nil {
		return err
	}
	roleArn, principleArn, err := saml.ParseArn(samlassertion)
	if err != nil {
		return fmt.Errorf("parsing rolearn and principle arn occurs error: %w", err)
	}
	//logger.Printf("role and priciple: \"%s\". \"%s\"", roleArn, principleArn)
	cred, err := aws.FetchCredentialViaSAML(ctx, *flag_region, roleArn, principleArn, samlresponse, *flag_duration)
	if err != nil {
		return err
	}
	aws.DumpCred(cred)
	return nil
}

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s (%s, %s, built at %s by %s):\n", os.Args[0], version, commit, date, builtBy)
		flag.PrintDefaults()
	}
	flag.Parse()
	logger := log.New(os.Stderr, "", log.LstdFlags)
	if err := fetchCreds(logger); err != nil {
		logger.Printf("occurs error: %s", err)
		os.Exit(1)
	}
}
