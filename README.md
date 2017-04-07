# sigctx
Signal for context.Context.

[![GoDoc](https://godoc.org/github.com/Code-Hex/sigctx?status.svg)](https://godoc.org/github.com/Code-Hex/sigctx) 
[![Build Status](https://travis-ci.org/Code-Hex/sigctx.svg?branch=master)](https://travis-ci.org/Code-Hex/sigctx) 
[![Go Report Card](https://goreportcard.com/badge/github.com/Code-Hex/sigctx)](https://goreportcard.com/report/github.com/Code-Hex/sigctx)

# Description
The package [sigctx](https://github.com/Code-Hex/sigctx) were developed to easily bind signals and context.Context.  
  
We use a combination of signals and context.Context often. It is mainly processing to context cancel after receiving the signal. Using this package makes it easy to do it.  
Example: [signal_cancel.go](https://github.com/Code-Hex/sigctx/blob/master/eg/signal_cancel.go)  

**By the way**, if you are using the context, have you ever wanted to send the signal to all the goroutines? You are lucky 🎉 With this package, it is possible to do something close to that.  
Example: [notify_signal.go](https://github.com/Code-Hex/sigctx/blob/master/eg/notify_signal.go)  
  
Do you want to run the above two methods at the same time? Of course you can do it.  
However, it becomes complicated if you mistake order. (Perhaps this problem may be fixed in the future.) But it will be very simple if the order of invocation is correct!!  
I recommend comparing [mistake_ctx_order.go](https://github.com/Code-Hex/sigctx/blob/master/eg/mistake_ctx_order.go) and [correct_ctx_order.go](https://github.com/Code-Hex/sigctx/blob/master/eg/correct_ctx_order.go).

# Install
    go get github.com/Code-Hex/sigctx

# Contributing
Welcome!!  
I'm waiting for pull requests or reporting issues.

# License
[MIT](https://github.com/Code-Hex/sigctx/blob/master/LICENSE)

# Author
[CodeHex](https://twitter.com/CodeHex)  
