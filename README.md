本程序用针对uTorrent实现自动屏蔽指定客户端的功能。
默认屏蔽、Xunlei、QQDownload、../torrent、aria2

=== 如图开启utorrent的web界面 ===
![image](https://github.com/tzwken/utorrent-block/assets/65268273/a9fd742d-dda5-4df0-9787-f449cdb81001)


```
程序参数说明：
      --key string    自定议额外要屏蔽的客户端关键字,正则表达式。默认屏蔽、Xunlei、QQDownload、../torrent、aria2
      --pass string   utorrent web页面的帐号
      --path string   指定uTorrent所在目录,将本程序放到uTorrent程序所在目录下可不配置此参数,可自行指定位置，例如'D:\utorrent'。
      --url string    uTorrent host default "http://127.0.0.1:1000/gui/"
      --user string   utorrent web页面的密码

程序运行说明：
    程序每30秒检测一次。每2小时会清空一次屏蔽的IP列表。当utorrent-block退出时会自动清理掉保存的IP信息。
```

程序运行截图

![image](https://github.com/tzwken/utorrent-block/assets/65268273/b83397a0-db56-484f-be96-ad000d3aa369)

推荐将程序放到utorrent所在目录，utorrent启动后，再双击 "运行utorrent-block.bat" 。通编辑此文件可修改帐号密码及额外要屏蔽的客户端信息。
