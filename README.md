[![Go Report Card](https://goreportcard.com/badge/github.com/yittg/ving)](https://goreportcard.com/report/github.com/yittg/ving)

# 🐸 ving

![](_assets/screenshot.png)

`ving` is a ping utility with nice output, in Golang(1.11+), enhanced with some useful features,
like trace, ports probe, and more yet to be implemented, 😝.

Special thanks to the amazing [termui](https://github.com/gizak/termui) library.

![](_assets/record.gif)

# 🦁 Features

* ping multiple targets concurrently and independently;
* trace a target like a simple `tracerout`, `--trace, -T`;
* probe well known tcp ports, `--ports`;
* error rate and latency statistics in sliding window, as emoji;
* sort by error rate and latency statistic, `--sort`;
* ping gateway conveniently, `-g`;
* responsive terminal display (based on termui).

## Feature details

| Features | Functionality | Details|
|----------|---------------|--------|
| Trace    | Toggle Key    | <kbd>t</kbd> |
|          | Switch        | <kbd>▲</kbd> / <kbd>k</kbd>, <kbd>▼</kbd> / <kbd>j</kbd> |
|          |               | <kbd>n</kbd>: manual mode |
|          |               | <kbd>c</kbd>: continuous mode |
| Ports    | Toggle Key    | <kbd>p</kbd> |
|          | Switch        | <kbd>▲</kbd> / <kbd>k</kbd>, <kbd>▼</kbd> / <kbd>j</kbd> |
|          |               | <kbd>f</kbd>: filter ports list, reached, unreached, or all |
|          |               | <kbd>v</kbd>: change view mode, name only, port number only, or both |
|          |               | <kbd>r</kbd>: refresh and probe all ports again |
| Help     | Toggle Key    | <kbd>h</kbd> |


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
$ ving

$ ving 8.8.8.8 -P 1-1024

$ ving --help
```
