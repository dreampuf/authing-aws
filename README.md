# Authing AWS

A CLI for fetching AWS Credentials on [Authing][5] via SAML Response.

# Install 

### Install via [Homebrew][3]

```shell
brew tap dreampuf/authing-aws
brew install authing-aws
```

### Install via [Docker][6]

```shell
docker pull dreampuf/authing-aws:latest
docker run -it --rm dreampuf/authing-aws:latest authing-aws #options
```

### Install manually

1. Find a usable binary in [release page][4]
2. Download binary and unarchive the package.
   ```shell
   curl -L -o ~/Downloads/authing-aws.tar.gz https://github.com/dreampuf/authing-aws/releases/download/v0.0.1/authing-aws_0.0.1_Darwin_x86_64.tar.gz
   tar -xf ~/Downloads/authing-aws.tar.gz
   ```

# Usage

You can get your AWS Credential in the following line:

```shell
authing-aws -url "https://path-to-authing-login-domain" \
  -username your-authing-username \
  -password your-password
  -app 0
```

Or to have a selected app

```shell
authing-aws -url "https://path-to-authing-login-domain" \
  -username your-authing-username \
  -password your-password
  -app "My AWS App"
```

You can set a function as a shortcut in your profile file 
```shell
aws-app () {
  eval $(authing-aws -url "https://path-to-authing-login-domain" \
    -username your-authing-username \
    -password your-password
    -app "My AWS App")
}
```

Test if it works
```shell
aws sts get-caller-identity
```

Help of `authing-aws`

```shell
authing-aws -help
Usage of authing-aws:
  -app string
    	selected app
  -debug
    	enable debug logs
  -disable-headless
    	disable headless mode to show chrome
  -duration int
    	duration in seconds (default 36000)
  -password string
    	password
  -region string
    	region of SAMLResponse (default "cn-north-1")
  -url string
    	URL
  -username string
    	username 
```

# How it works

It capture the [SAML Response](https://docs.amazonaws.cn/en_us/IAM/latest/UserGuide/troubleshoot_saml_view-saml-response.html) and use it to fetch aws access token.  
This depends on [chromedp][1] to communicate with browser (Chrome) and [aws-sdk-go][2] to exchange access token.

# License 

GPL-3.0 license

[1]: https://github.com/chromedp/chromedp
[2]: https://github.com/aws/aws-sdk-go
[3]: https://brew.sh/
[4]: https://github.com/dreampuf/authing-aws/releases
[5]: https://www.authing.cn/