# Authing AWS

A shortcut for fetching AWS Credentials via SAML Response.

# Usage




# How it works

It capture the [SAML Response](https://docs.amazonaws.cn/en_us/IAM/latest/UserGuide/troubleshoot_saml_view-saml-response.html) and use it to fetch aws access token.  
This depends on [chromedp][1] to communicate with browser (Chrome) and [aws-sdk-go][2] to exchange access token.

[1]: https://github.com/chromedp/chromedp
[2]: https://github.com/aws/aws-sdk-go