config = {
    "novel_dirname": "~/.novel/novels/json",
    "novel_extname": ".novel",
    "icon_dirname": "~/.novel/icons",
    "icon_extname": ".img",
    "log_dirname": "~/.novel/log",
    "log_filename": "novel.log",
    "max_log_file_size": 10 * 1024 * 1024,
    "log_backup_count": 5,
    "ignore_host_file": ".ignored_host_file" #不能变，写死的
}

# 用户只要直接修改上面的数据的值就可以，并且支持~, 和$HOME
# 下面将值中的~和$HOME替换为当前的用户名称
import os

home_dir = os.environ.get('HOME')
for key in config:
    if isinstance(config[key], str):
        config[key] = config[key].replace("~", home_dir)
        config[key] = config[key].replace("$HOME", home_dir)

if not os.path.exists(config['log_dirname']):
    os.makedirs(config['log_dirname'])
