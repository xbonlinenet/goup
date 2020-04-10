

### 创建工程


**linux**

`bash  -c "$(curl -fsSL https://raw.githubusercontent.com/xbonlinenet/goup/master/init.sh)"`

**mac**

mac 对 shell 的支持和 linux 的支持有差异，暂时不兼容。
建议修改 mac 的环境：

mac 安装gnu的sed命令，这样就可以跟linux环境兼容了
1. brew install gnu-sed
2.

    ```
    vi ~/.zshrc
    export PATH="/usr/local/opt/gnu-sed/libexec/gnubin:$PATH"
    ```
3. source ~/.zshrc 或者新开窗口，让设置生效
