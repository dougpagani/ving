[![Go Report Card](https://goreportcard.com/badge/github.com/yittg/ving)](https://goreportcard.com/report/github.com/yittg/ving)

# 🐸 ving

`ving` is a visualization ping utility written in Golang(1.11+).
Special thanks to the amazing [termui](https://github.com/gizak/termui) library.

# 🦁 Features

* ping multiple targets concurrently and independently;
* trace a target like a simple `tracerout`;
* error rate and latency statistics in sliding window, as emoji;
* sort by error rate and latency statistic, `--sort`;
* ping gateway conveniently, `-g`;
* responsive terminal display (based on termui).

## Feature details

### trace

toggle by tapping <kbd>t</kbd> and use <kbd>▲</kbd> / <kbd>▼</kbd> to choose then <kbd>enter</kbd> to begin a trace.

By default, `ving` will trace a target automatically per 500ms along with the TTL increase until it touches the target, then begin another round trace.
You can enable trace manually by tapping <kbd>n</kbd>, and each following tapping <kbd>n</kbd> will increase ttl and do a single probe.
Absolutely, you can also recover to automatic mode by tapping <kbd>c</kbd>.


# 🙈 Install

```
$ go get -u github.com/yittg/ving
```

> __Notes__ for linux users, run `ving` with `sudo` or `setcap` in advance, 
for more information, see the [man page](http://linux.die.net/man/7/capabilities).
>
>    ```
>    $ sudo setcap "cap_net_raw+ep" ving
>    ``` 

# ⚡ Usage

```
$ ving 192.168.0.1 127.0.0.1 8.8.8.8

$ ving -i 100ms 192.168.0.1

$ ving -g
```

![](./assets/screenshot.png)
