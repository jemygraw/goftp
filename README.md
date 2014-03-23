#goftp

##背景
该项目由多科学堂的小伙伴@Tony想出来的，希望大家能够通过实现一个小的项目来学习go语言。之所以选择实现ftp客户端是因为@Jemy同学提议的，而且由于ftp协议的相对简单，也可以保证每一个小伙伴都能够从中学习一点东西。

##参与者
1. 多科学堂的所有小伙伴们
2. 所有爱好学习的小伙伴们

##发起时间
2014年3月14日

##简单介绍
本项目主要以实现Windows下面的命令ftp为先。然后可能扩展到linux下面的命令ftp，因为linux下面的ftp命令太多，所以开始的时候以简单的为主。
Windows下面的ftp命令主要有如下这些：

|    命令   |             原型                  |描述                                     |
|----------|-----------------------------------|----------------------------------------|
|!         |!                                  |打印版权信息,并退出程序                     |
|?         |? [cmd_name]                       |打印所有命令或者指定命令的帮助信息，功能同help |
|append    |append local_file [remote_file]    |追加文件内容                              |
|ascii     |ascii                              |设置文件传输模式为文本模式                   |
|bell      |bell                               |设置命令结束后响铃                          |
|binary    |binary                             |设置文件传输模式为二进制模式                  |
|bye       |bye                                |关闭ftp连接并退出程序                       |
|cd        |cd remote_dir                      |切换工作路径                               |
|close     |close                              |关闭ftp连接                               |
|delete    |delete remote_file                 |删除远程文件                               |
|debug     |debug                              |打开或关闭调试模式                          |
|dir       |dir [remote_dir][local_file]       |打印远程目录的目录详细内容,包括文件夹和文件以及`.`和`..`, 并可以将结果另存为文件     |
|disconnect|disconnect                         |关闭ftp连接，功能同close                    |
|get       |get remote_file [local_file]       |获取远程文件，并可以另存为另一个文件           |
|glob      |glob                               ||
|hash      |hash                               ||
|help      |help [cmd_name]                    |打印所有命令或者指定命令的帮助信息，功能同?      |
|lcd       |lcd [local_dir]                    |切换本地工作目录,默认路径为启动ftp命令的目录    |
|literal   |literal argument [...]             |将命令参数逐个发送给远程服务器，远程服务器逐个响应|
|ls        |ls [remote_dir][local_file]        |打印远程目录的文件列表，仅包括文件，可以将结果保存在文件中|
|mdelete   |mdelete remote_file [...]          |删除多个远程文件，支持通配符                  |
|mdir      |mdir remote_dir [...] [local_file] |打印多个远程目录的详细内容，并可以保存在本地文件中|





##测试环境
1. 在Windows的环境下，大家可以下载一个Server U的ftp服务器软件，然后配置一样，用来测试。
2. Mac下面，如果不想折腾，装个Windows虚拟机吧。

##已经实现命令列表
?
bye
cd
close
help
quit
version
disconnect
pwd
lcd
ls
open


