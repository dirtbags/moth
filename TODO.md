* Figure out how to log JSend short text in addition to HTTP code
* We've got logic in state.go and httpd.go that is neither httpd nor state specific.
  Pull this into some other file that means "here are the brains of the server".
