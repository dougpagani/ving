# ving
`ving` is a visual ping utility written in Golang.
Special thanks [termui](https://github.com/gizak/termui) for providing terminal library.

# Features

* ping multiple targets independently at once;
* ping interval option.

# Usage

```
$ ving 192.168.0.1 127.0.0.1 8.8.8.8

$ ving -i 100ms 192.168.0.1
```

![](./assets/screenshot.png)